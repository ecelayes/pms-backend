package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ecelayes/pms-backend/internal/shared/domain"
	"github.com/redis/go-redis/v9"
)

// AnalyticsEvent is the public struct stored in the analytics stream.
type AnalyticsEvent = domain.AnalyticsEvent

// AnalyticsStore implements domain.AnalyticsRecorder using Redis.
type AnalyticsStore struct {
	client *redis.Client
}

var _ domain.AnalyticsRecorder = (*AnalyticsStore)(nil)

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
	case domain.ReservationCreatedPayload:
		event.PropertyID = p.PropertyID
		event.UnitTypeID = p.UnitTypeID
		event.GuestEmail = p.GuestEmail
		event.Data = map[string]interface{}{
			"reservation_code": p.ReservationCode,
			"start_date":       p.StartDate,
			"end_date":         p.EndDate,
		}
	case domain.ReservationConfirmedPayload:
		event.PropertyID = p.PropertyID
		event.GuestEmail = p.GuestEmail
		event.Data = map[string]interface{}{
			"reservation_code": p.ReservationCode,
		}
	case domain.ReservationCancelledPayload:
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
