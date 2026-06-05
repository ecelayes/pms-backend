package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ecelayes/pms-backend/internal/shared/domain"
	"github.com/redis/go-redis/v9"
)

// StreamMessage is the message envelope used by StreamProducer/Consumer.
type StreamMessage = domain.StreamMessage

// Stream/group names. These are implementation details of the Redis adapter
// and intentionally NOT exposed via the domain ports.
const (
	StreamReservations = "reservation:events"
	StreamDLQ          = "reservation:events:dlq"
	GroupAvailability  = "availability-group"
	GroupNotifications = "notifications-group"
	GroupAnalytics     = "analytics-group"
	MaxRetries         = 3
)

// StreamProducer implements domain.StreamProducer using Redis Streams.
type StreamProducer struct {
	client *redis.Client
}

var _ domain.StreamProducer = (*StreamProducer)(nil)

func NewStreamProducer(client *redis.Client) *StreamProducer {
	return &StreamProducer{client: client}
}

func (p *StreamProducer) PublishReservationCreated(ctx context.Context, payload domain.ReservationCreatedPayload) error {
	return p.addToStream(ctx, domain.EventReservationCreated, payload)
}
func (p *StreamProducer) PublishReservationConfirmed(ctx context.Context, payload domain.ReservationConfirmedPayload) error {
	return p.addToStream(ctx, domain.EventReservationConfirmed, payload)
}
func (p *StreamProducer) PublishReservationCancelled(ctx context.Context, payload domain.ReservationCancelledPayload) error {
	return p.addToStream(ctx, domain.EventReservationCancelled, payload)
}

func (p *StreamProducer) addToStream(ctx context.Context, eventType string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	_, err = p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: StreamReservations,
		Values: map[string]interface{}{
			"event_type": eventType,
			"payload":    string(payloadBytes),
			"created_at": time.Now().Unix(),
		},
	}).Result()
	if err != nil {
		return fmt.Errorf("failed to XADD to stream: %w", err)
	}
	log.Printf("[StreamProducer] Added event %s to stream %s", eventType, StreamReservations)
	return nil
}

func (p *StreamProducer) EnsureGroups(ctx context.Context) error {
	groups := []string{GroupAvailability, GroupNotifications, GroupAnalytics}
	for _, group := range groups {
		err := p.client.XGroupCreateMkStream(ctx, StreamReservations, group, "0").Err()
		if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
			return fmt.Errorf("failed to create group %s: %w", group, err)
		}
	}
	log.Printf("[StreamProducer] Ensured all consumer groups exist")
	return nil
}

// StreamConsumer implements domain.StreamConsumer using Redis Streams.
type StreamConsumer struct {
	client       *redis.Client
	groupName    string
	consumerName string
}

var _ domain.StreamConsumer = (*StreamConsumer)(nil)

func NewStreamConsumer(client *redis.Client, groupName, consumerName string) *StreamConsumer {
	return &StreamConsumer{
		client:       client,
		groupName:    groupName,
		consumerName: consumerName,
	}
}

func (c *StreamConsumer) Consume(ctx context.Context, eventTypes []string, handler domain.StreamEventHandler) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("[StreamConsumer] Context cancelled, stopping consumer %s", c.consumerName)
			return
		default:
			c.readAndProcess(ctx, eventTypes, handler)
		}
	}
}

func (c *StreamConsumer) readAndProcess(ctx context.Context, eventTypes []string, handler domain.StreamEventHandler) {
	streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    c.groupName,
		Consumer: c.consumerName,
		Streams:  []string{StreamReservations, ">"},
		Count:    1,
		Block:    2 * time.Second,
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return
		}
		log.Printf("[StreamConsumer] XReadGroup error: %v", err)
		time.Sleep(100 * time.Millisecond)
		return
	}
	for _, stream := range streams {
		for _, message := range stream.Messages {
			eventType := message.Values["event_type"].(string)
			if len(eventTypes) > 0 && !contains(eventTypes, eventType) {
				c.ackMessage(ctx, message.ID)
				continue
			}
			payload := []byte(message.Values["payload"].(string))
			msg := &StreamMessage{
				ID:        message.ID,
				EventType: eventType,
				Payload:   payload,
				Retries:   0,
			}
			if err := handler(msg); err != nil {
				log.Printf("[StreamConsumer] Handler error for %s: %v", msg.ID, err)
				retryStr, ok := message.Values["retry_count"].(string)
				if ok {
					retryCount, _ := strconv.Atoi(retryStr)
					msg.Retries = retryCount
				}
				if msg.Retries >= MaxRetries {
					c.moveToDLQ(ctx, msg)
					c.ackMessage(ctx, message.ID)
					log.Printf("[StreamConsumer] Message %s moved to DLQ after %d retries", msg.ID, msg.Retries)
				} else {
					c.nackWithRetry(ctx, msg)
				}
			} else {
				c.ackMessage(ctx, message.ID)
			}
		}
	}
}

func (c *StreamConsumer) ackMessage(ctx context.Context, messageID string) {
	err := c.client.XAck(ctx, StreamReservations, c.groupName, messageID).Err()
	if err != nil {
		log.Printf("[StreamConsumer] XACK error for %s: %v", messageID, err)
	}
}

func (c *StreamConsumer) nackWithRetry(ctx context.Context, msg *StreamMessage) {
	currentRetry := msg.Retries
	newRetry := currentRetry + 1
	_, err := c.client.XClaim(ctx, &redis.XClaimArgs{
		Stream:   StreamReservations,
		Group:    c.groupName,
		Consumer: c.consumerName,
		MinIdle:  30 * time.Second,
		Messages: []string{msg.ID},
	}).Result()
	if err != nil {
		log.Printf("[StreamConsumer] XCLAIM error for %s: %v", msg.ID, err)
	} else {
		log.Printf("[StreamConsumer] Requeued message %s (retry %d)", msg.ID, newRetry)
	}
}

func (c *StreamConsumer) moveToDLQ(ctx context.Context, msg *StreamMessage) {
	_, err := c.client.XAdd(ctx, &redis.XAddArgs{
		Stream: StreamDLQ,
		Values: map[string]interface{}{
			"original_id": msg.ID,
			"event_type":  msg.EventType,
			"payload":     string(msg.Payload),
			"failed_at":   time.Now().Unix(),
			"retries":     msg.Retries,
		},
	}).Result()
	if err != nil {
		log.Printf("[StreamConsumer] Failed to add message to DLQ: %v", err)
	}
}

func contains(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
