package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type AnalyticsEvent struct {
	ID         string                 `json:"id"`
	EventType  string                 `json:"event_type"`
	PropertyID string                 `json:"property_id"`
	UnitTypeID string                 `json:"unit_type_id"`
	GuestEmail string                 `json:"guest_email"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

type AnalyticsStore struct {
	client *redis.Client
}

func NewAnalyticsStore(client *redis.Client) *AnalyticsStore {
	return &AnalyticsStore{client: client}
}

func (s *AnalyticsStore) RecordEvent(ctx context.Context, eventType string, payload interface{}) error {
	event := AnalyticsEvent{
		ID:         fmt.Sprintf("%d-%d", time.Now().UnixMilli(), time.Now().UnixNano()%1000),
		EventType:  eventType,
		Timestamp: time.Now(),
	}
	
	switch p := payload.(type) {
	case ReservationCreatedPayload:
		event.PropertyID = p.PropertyID
		event.UnitTypeID = p.UnitTypeID
		event.GuestEmail = p.GuestEmail
		event.Data = map[string]interface{}{
			"reservation_code": p.ReservationCode,
			"start_date":       p.StartDate,
			"end_date":         p.EndDate,
		}
	case ReservationConfirmedPayload:
		event.PropertyID = p.PropertyID
		event.GuestEmail = p.GuestEmail
		event.Data = map[string]interface{}{
			"reservation_code": p.ReservationCode,
		}
	case ReservationCancelledPayload:
		event.PropertyID = p.PropertyID
		event.GuestEmail = p.GuestEmail
		event.Data = map[string]interface{}{
			"reservation_code": p.ReservationCode,
		}
	}
	
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal analytics event: %w", err)
	}
	
	dateKey := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("analytics:%s:%s", eventType, dateKey)
	return s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: key,
		Values: map[string]interface{}{
			"event": string(data),
		},
	}).Err()
}

func (s *AnalyticsStore) GetEventCounts(ctx context.Context, days int) (map[string]int64, error) {
	counts := make(map[string]int64)
	now := time.Now()
	
	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i)
		key := fmt.Sprintf("analytics:reservation.created:%s", date.Format("2006-01-02"))
		
		count, err := s.client.XLen(ctx, key).Result()
		if err != nil && err != redis.Nil {
			continue
		}
		counts[date.Format("2006-01-02")] = count
	}
	
	return counts, nil
}
