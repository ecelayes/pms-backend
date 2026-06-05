package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ecelayes/pms-backend/internal/shared/domain"
	"github.com/redis/go-redis/v9"
)

// EventPublisher implements domain.EventPublisher using Redis Pub/Sub.
type EventPublisher struct {
	client *redis.Client
}

// Verify interface compliance
var _ domain.EventPublisher = (*EventPublisher)(nil)

func NewEventPublisher(client *redis.Client) *EventPublisher {
	return &EventPublisher{client: client}
}

func (p *EventPublisher) PublishReservationCreated(ctx context.Context, payload domain.ReservationCreatedPayload) error {
	event := map[string]interface{}{
		"Name":       domain.EventReservationCreated,
		"OccurredAt": time.Now(),
		"Data":       payload,
	}
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	channel := "events:" + domain.EventReservationCreated
	return p.client.Publish(ctx, channel, payloadBytes).Err()
}

func (p *EventPublisher) PublishReservationConfirmed(ctx context.Context, payload domain.ReservationConfirmedPayload) error {
	event := map[string]interface{}{
		"Name":       domain.EventReservationConfirmed,
		"OccurredAt": time.Now(),
		"Data":       payload,
	}
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	channel := "events:" + domain.EventReservationConfirmed
	return p.client.Publish(ctx, channel, payloadBytes).Err()
}

func (p *EventPublisher) PublishReservationCancelled(ctx context.Context, payload domain.ReservationCancelledPayload) error {
	event := map[string]interface{}{
		"Name":       domain.EventReservationCancelled,
		"OccurredAt": time.Now(),
		"Data":       payload,
	}
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	channel := "events:" + domain.EventReservationCancelled
	return p.client.Publish(ctx, channel, payloadBytes).Err()
}
