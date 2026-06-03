package domain

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

const (
	EventChannelPrefix = "events:"
)

type RedisPublisher struct {
	client *redis.Client
}

func NewRedisPublisher(client *redis.Client) *RedisPublisher {
	return &RedisPublisher{client: client}
}
func (p *RedisPublisher) PublishReservationCreated(ctx context.Context, payload ReservationCreatedPayload) error {
	event := map[string]interface{}{
		"Name":       EventReservationCreated,
		"OccurredAt": time.Now(),
		"Data":       payload,
	}
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	channel := EventChannelPrefix + EventReservationCreated
	return p.client.Publish(ctx, channel, payloadBytes).Err()
}
func (p *RedisPublisher) PublishReservationConfirmed(ctx context.Context, payload ReservationConfirmedPayload) error {
	event := map[string]interface{}{
		"Name":       EventReservationConfirmed,
		"OccurredAt": time.Now(),
		"Data":       payload,
	}
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	channel := EventChannelPrefix + EventReservationConfirmed
	return p.client.Publish(ctx, channel, payloadBytes).Err()
}
func (p *RedisPublisher) PublishReservationCancelled(ctx context.Context, payload ReservationCancelledPayload) error {
	event := map[string]interface{}{
		"Name":       EventReservationCancelled,
		"OccurredAt": time.Now(),
		"Data":       payload,
	}
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	channel := EventChannelPrefix + EventReservationCancelled
	return p.client.Publish(ctx, channel, payloadBytes).Err()
}

type EventSubscriber struct {
	client *redis.Client
}

func NewEventSubscriber(client *redis.Client) *EventSubscriber {
	return &EventSubscriber{client: client}
}
func (s *EventSubscriber) SubscribeReservationCreated(ctx context.Context, handler func(payload ReservationCreatedPayload)) {
	channel := EventChannelPrefix + EventReservationCreated
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
			eventName, ok := event["Name"].(string)
			if !ok || eventName != EventReservationCreated {
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
			var payload ReservationCreatedPayload
			if err := json.Unmarshal(dataBytes, &payload); err != nil {
				log.Printf("[EventSubscriber] Failed to unmarshal reservation payload: %v", err)
				continue
			}
			handler(payload)
		}
	}
}
func (s *EventSubscriber) SubscribeReservationCancelled(ctx context.Context, handler func(payload ReservationCancelledPayload)) {
	channel := EventChannelPrefix + EventReservationCancelled
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
			eventName, ok := event["Name"].(string)
			if !ok || eventName != EventReservationCancelled {
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
			var payload ReservationCancelledPayload
			if err := json.Unmarshal(dataBytes, &payload); err != nil {
				log.Printf("[EventSubscriber] Failed to unmarshal reservation payload: %v", err)
				continue
			}
			handler(payload)
		}
	}
}
