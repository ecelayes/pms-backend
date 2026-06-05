package redis

import (
	"context"
	"encoding/json"
	"log"

	"github.com/ecelayes/pms-backend/internal/shared/domain"
	"github.com/redis/go-redis/v9"
)

// EventSubscriber implements domain.EventSubscriber using Redis Pub/Sub.
type EventSubscriber struct {
	client *redis.Client
}

// Verify interface compliance
var _ domain.EventSubscriber = (*EventSubscriber)(nil)

func NewEventSubscriber(client *redis.Client) *EventSubscriber {
	return &EventSubscriber{client: client}
}

func (s *EventSubscriber) SubscribeReservationCreated(ctx context.Context, handler func(payload domain.ReservationCreatedPayload)) {
	s.subscribe(ctx, domain.EventReservationCreated, func(data []byte) {
		var payload domain.ReservationCreatedPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			log.Printf("[EventSubscriber] Failed to unmarshal reservation payload: %v", err)
			return
		}
		handler(payload)
	})
}

func (s *EventSubscriber) SubscribeReservationCancelled(ctx context.Context, handler func(payload domain.ReservationCancelledPayload)) {
	s.subscribe(ctx, domain.EventReservationCancelled, func(data []byte) {
		var payload domain.ReservationCancelledPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			log.Printf("[EventSubscriber] Failed to unmarshal reservation payload: %v", err)
			return
		}
		handler(payload)
	})
}

// subscribe is the shared implementation for all event subscriptions.
func (s *EventSubscriber) subscribe(ctx context.Context, eventName string, dispatch func(data []byte)) {
	channel := "events:" + eventName
	pubsub := s.client.Subscribe(ctx, channel)
	defer pubsub.Close()
	ch := pubsub.Channel()
	log.Printf("[EventSubscriber] Listening on channel: %s", channel)
	for {
		select {
		case <-ctx.Done():
			log.Printf("[EventSubscriber] Context cancelled, stopping subscription")
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}
			var event map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("[EventSubscriber] Failed to unmarshal event: %v", err)
				continue
			}
			envelopeName, ok := event["Name"].(string)
			if !ok || envelopeName != eventName {
				continue
			}
			dataField, ok := event["Data"].(map[string]interface{})
			if !ok {
				log.Printf("[EventSubscriber] Event data is not a map")
				continue
			}
			dataBytes, err := json.Marshal(dataField)
			if err != nil {
				log.Printf("[EventSubscriber] Failed to marshal event data: %v", err)
				continue
			}
			dispatch(dataBytes)
		}
	}
}
