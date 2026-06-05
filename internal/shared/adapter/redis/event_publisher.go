package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ecelayes/pms-backend/internal/shared/domain"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// EventPublisher implements domain.EventPublisher using Redis Pub/Sub.
type EventPublisher struct {
	client *redis.Client
	logger *zap.Logger
}

// Verify interface compliance
var _ domain.EventPublisher = (*EventPublisher)(nil)

func NewEventPublisher(client *redis.Client, logger *zap.Logger) *EventPublisher {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &EventPublisher{client: client, logger: logger}
}

func (p *EventPublisher) PublishReservationCreated(ctx context.Context, payload domain.ReservationCreatedPayload) error {
	return p.publish(ctx, domain.EventReservationCreated, payload)
}

func (p *EventPublisher) PublishReservationConfirmed(ctx context.Context, payload domain.ReservationConfirmedPayload) error {
	return p.publish(ctx, domain.EventReservationConfirmed, payload)
}

func (p *EventPublisher) PublishReservationCancelled(ctx context.Context, payload domain.ReservationCancelledPayload) error {
	return p.publish(ctx, domain.EventReservationCancelled, payload)
}

func (p *EventPublisher) publish(ctx context.Context, eventName string, data interface{}) error {
	event := map[string]interface{}{
		"Name":       eventName,
		"OccurredAt": time.Now(),
		"Data":       data,
	}
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event %s: %w", eventName, err)
	}
	channel := EventChannelPrefix + eventName
	_, err = withRetry(ctx, defaultRetryConfig(), func(ctx context.Context) (struct{}, error) {
		return struct{}{}, p.client.Publish(ctx, channel, payloadBytes).Err()
	})
	if err != nil {
		p.logger.Error("failed to publish event",
			zap.String("event", eventName),
			zap.String("channel", channel),
			zap.Error(err))
		return err
	}
	p.logger.Debug("published event",
		zap.String("event", eventName),
		zap.String("channel", channel))
	return nil
}
