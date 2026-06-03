package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"strconv"
	"time"
)

const (
	StreamReservations = "reservation:events"
	StreamDLQ          = "reservation:events:dlq"
	GroupAvailability  = "availability-group"
	GroupNotifications = "notifications-group"
	GroupAnalytics     = "analytics-group"
	MaxRetries         = 3
)

type StreamMessage struct {
	ID        string
	EventType string
	Payload   []byte
	Retries   int
}
type RedisStreamProducer struct {
	client *redis.Client
}

func NewRedisStreamProducer(client *redis.Client) *RedisStreamProducer {
	return &RedisStreamProducer{client: client}
}
func (p *RedisStreamProducer) PublishReservationCreated(ctx context.Context, payload ReservationCreatedPayload) error {
	return p.addToStream(ctx, EventReservationCreated, payload)
}
func (p *RedisStreamProducer) PublishReservationConfirmed(ctx context.Context, payload ReservationConfirmedPayload) error {
	return p.addToStream(ctx, EventReservationConfirmed, payload)
}
func (p *RedisStreamProducer) PublishReservationCancelled(ctx context.Context, payload ReservationCancelledPayload) error {
	return p.addToStream(ctx, EventReservationCancelled, payload)
}
func (p *RedisStreamProducer) addToStream(ctx context.Context, eventType string, payload interface{}) error {
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
	log.Printf("[RedisStreamProducer] Added event %s to stream %s", eventType, StreamReservations)
	return nil
}
func (p *RedisStreamProducer) EnsureGroups(ctx context.Context) error {
	groups := []string{GroupAvailability, GroupNotifications, GroupAnalytics}
	for _, group := range groups {
		err := p.client.XGroupCreateMkStream(ctx, StreamReservations, group, "0").Err()
		if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
			return fmt.Errorf("failed to create group %s: %w", group, err)
		}
	}
	log.Printf("[RedisStreamProducer] Ensured all consumer groups exist")
	return nil
}

type RedisStreamConsumer struct {
	client       *redis.Client
	groupName    string
	consumerName string
}

func NewRedisStreamConsumer(client *redis.Client, groupName, consumerName string) *RedisStreamConsumer {
	return &RedisStreamConsumer{
		client:       client,
		groupName:    groupName,
		consumerName: consumerName,
	}
}

type StreamEventHandler func(msg *StreamMessage) error

func (c *RedisStreamConsumer) Consume(ctx context.Context, eventTypes []string, handler StreamEventHandler) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("[RedisStreamConsumer] Context cancelled, stopping consumer %s", c.consumerName)
			return
		default:
			c.readAndProcess(ctx, eventTypes, handler)
		}
	}
}
func (c *RedisStreamConsumer) readAndProcess(ctx context.Context, eventTypes []string, handler StreamEventHandler) {
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
		log.Printf("[RedisStreamConsumer] XReadGroup error: %v", err)
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
				log.Printf("[RedisStreamConsumer] Handler error for %s: %v", msg.ID, err)
				retryStr, ok := message.Values["retry_count"].(string)
				if ok {
					retryCount, _ := strconv.Atoi(retryStr)
					msg.Retries = retryCount
				}
				if msg.Retries >= MaxRetries {
					c.moveToDLQ(ctx, msg)
					c.ackMessage(ctx, message.ID)
					log.Printf("[RedisStreamConsumer] Message %s moved to DLQ after %d retries", msg.ID, msg.Retries)
				} else {
					c.nackWithRetry(ctx, msg)
				}
			} else {
				c.ackMessage(ctx, message.ID)
			}
		}
	}
}
func (c *RedisStreamConsumer) ackMessage(ctx context.Context, messageID string) {
	err := c.client.XAck(ctx, StreamReservations, c.groupName, messageID).Err()
	if err != nil {
		log.Printf("[RedisStreamConsumer] XACK error for %s: %v", messageID, err)
	}
}
func (c *RedisStreamConsumer) nackWithRetry(ctx context.Context, msg *StreamMessage) {
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
		log.Printf("[RedisStreamConsumer] XCLAIM error for %s: %v", msg.ID, err)
	} else {
		log.Printf("[RedisStreamConsumer] Requeued message %s (retry %d)", msg.ID, newRetry)
	}
}
func (c *RedisStreamConsumer) moveToDLQ(ctx context.Context, msg *StreamMessage) {
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
		log.Printf("[RedisStreamConsumer] Failed to add message to DLQ: %v", err)
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

type StreamPublisher interface {
	PublishReservationCreated(ctx context.Context, payload ReservationCreatedPayload) error
	PublishReservationConfirmed(ctx context.Context, payload ReservationConfirmedPayload) error
	PublishReservationCancelled(ctx context.Context, payload ReservationCancelledPayload) error
}
