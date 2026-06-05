package redis

import (
	"context"
	"encoding/json"

	"github.com/ecelayes/pms-backend/internal/shared/domain"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// EventSubscriber implements domain.EventSubscriber using Redis Pub/Sub.
type EventSubscriber struct {
	client *redis.Client
	logger *zap.Logger
}

// Verify interface compliance
var _ domain.EventSubscriber = (*EventSubscriber)(nil)

func NewEventSubscriber(client *redis.Client, logger *zap.Logger) *EventSubscriber {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &EventSubscriber{client: client, logger: logger}
}

func (s *EventSubscriber) SubscribeReservationCreated(ctx context.Context, handler func(payload domain.ReservationCreatedPayload)) {
	s.subscribe(ctx, domain.EventReservationCreated, func(data []byte) {
		var payload domain.ReservationCreatedPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			s.logger.Error("failed to unmarshal reservation payload",
				zap.String("event", domain.EventReservationCreated),
				zap.Error(err))
			return
		}
		handler(payload)
	})
}

func (s *EventSubscriber) SubscribeReservationCancelled(ctx context.Context, handler func(payload domain.ReservationCancelledPayload)) {
	s.subscribe(ctx, domain.EventReservationCancelled, func(data []byte) {
		var payload domain.ReservationCancelledPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			s.logger.Error("failed to unmarshal reservation payload",
				zap.String("event", domain.EventReservationCancelled),
				zap.Error(err))
			return
		}
		handler(payload)
	})
}

// subscribe is the shared implementation for all event subscriptions.
func (s *EventSubscriber) subscribe(ctx context.Context, eventName string, dispatch func(data []byte)) {
	channel := EventChannelPrefix + eventName
	pubsub := s.client.Subscribe(ctx, channel)
	defer func() { _ = pubsub.Close() }()
	ch := pubsub.Channel()
	s.logger.Info("listening on channel", zap.String("channel", channel))
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("context cancelled, stopping subscription", zap.String("channel", channel))
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}
			var event map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				s.logger.Error("failed to unmarshal event",
					zap.String("channel", channel),
					zap.Error(err))
				continue
			}
			envelopeName, ok := event["Name"].(string)
			if !ok || envelopeName != eventName {
				continue
			}
			dataField, ok := event["Data"].(map[string]interface{})
			if !ok {
				s.logger.Warn("event data is not a map",
					zap.String("channel", channel),
					zap.Any("event", event))
				continue
			}
			dataBytes, err := json.Marshal(dataField)
			if err != nil {
				s.logger.Error("failed to marshal event data",
					zap.String("channel", channel),
					zap.Error(err))
				continue
			}
			dispatch(dataBytes)
		}
	}
}
