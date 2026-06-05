package redis

import (
	"context"
	"errors"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ecelayes/pms-backend/internal/shared/domain"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// EventChannelPrefix is the prefix used for Redis Pub/Sub event channels.
const EventChannelPrefix = "events:"

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

// StreamMessage is the message envelope used by StreamProducer/Consumer.
type StreamMessage = domain.StreamMessage

// StreamProducer implements domain.StreamProducer using Redis Streams.
type StreamProducer struct {
	client *redis.Client
	logger *zap.Logger
}

var _ domain.StreamProducer = (*StreamProducer)(nil)

func NewStreamProducer(client *redis.Client, logger *zap.Logger) *StreamProducer {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &StreamProducer{client: client, logger: logger}
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
	p.logger.Info("added event to stream",
		zap.String("event_type", eventType),
		zap.String("stream", StreamReservations))
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
	p.logger.Info("ensured all consumer groups exist",
		zap.String("stream", StreamReservations),
		zap.Strings("groups", groups))
	return nil
}

// StreamConsumer implements domain.StreamConsumer using Redis Streams.
type StreamConsumer struct {
	client       *redis.Client
	groupName    string
	consumerName string
	logger       *zap.Logger
}

var _ domain.StreamConsumer = (*StreamConsumer)(nil)

func NewStreamConsumer(client *redis.Client, groupName, consumerName string, logger *zap.Logger) *StreamConsumer {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &StreamConsumer{
		client:       client,
		groupName:    groupName,
		consumerName: consumerName,
		logger:       logger,
	}
}

func (c *StreamConsumer) Consume(ctx context.Context, eventTypes []string, handler domain.StreamEventHandler) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("context cancelled, stopping consumer",
				zap.String("consumer", c.consumerName),
				zap.String("group", c.groupName))
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
		if errors.Is(err, redis.Nil) {
			return
		}
		c.logger.Error("XReadGroup error",
			zap.String("consumer", c.consumerName),
			zap.Error(err))
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
				c.logger.Error("handler error",
					zap.String("message_id", msg.ID),
					zap.String("event_type", eventType),
					zap.Error(err))
				retryStr, ok := message.Values["retry_count"].(string)
				if ok {
					retryCount, _ := strconv.Atoi(retryStr)
					msg.Retries = retryCount
				}
				if msg.Retries >= MaxRetries {
					c.moveToDLQ(ctx, msg)
					c.ackMessage(ctx, message.ID)
					c.logger.Warn("message moved to DLQ",
						zap.String("message_id", msg.ID),
						zap.Int("retries", msg.Retries))
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
		c.logger.Error("XACK error",
			zap.String("message_id", messageID),
			zap.Error(err))
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
		c.logger.Error("XCLAIM error",
			zap.String("message_id", msg.ID),
			zap.Error(err))
	} else {
		c.logger.Info("message requeued",
			zap.String("message_id", msg.ID),
			zap.Int("retry", newRetry))
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
		c.logger.Error("failed to add message to DLQ",
			zap.String("message_id", msg.ID),
			zap.Error(err))
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
