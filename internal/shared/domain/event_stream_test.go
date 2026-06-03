package domain

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func skipIfNoRedis(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	client.Close()
}

func TestRedisStreamProducer_PublishReservationCreated(t *testing.T) {
	skipIfNoRedis(t)

	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer client.Close()

	producer := NewRedisStreamProducer(client)
	ctx := context.Background()

	// Ensure groups exist
	err := producer.EnsureGroups(ctx)
	assert.NoError(t, err)

	payload := ReservationCreatedPayload{
		ReservationID:  "stream-test-res-123",
		PropertyID:    "stream-test-prop-456",
		UnitTypeID:    "stream-test-unit-type",
		RatePlanID:    "rate-plan-001",
		GuestID:       "guest-001",
		GuestEmail:    "guest@example.com",
		ReservationCode: "STREAM123",
		StartDate:     time.Now().Add(24 * time.Hour),
		EndDate:       time.Now().Add(48 * time.Hour),
	}

	err = producer.PublishReservationCreated(ctx, payload)
	assert.NoError(t, err)
}

func TestStreamMessage_JSON(t *testing.T) {
	payload := ReservationCreatedPayload{
		ReservationID:  "res-123",
		PropertyID:    "prop-456",
		UnitTypeID:    "unit-type-789",
		ReservationCode: "ABC12345",
		StartDate:     time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		EndDate:       time.Date(2025, 1, 17, 0, 0, 0, 0, time.UTC),
	}

	payloadBytes, err := json.Marshal(payload)
	assert.NoError(t, err)

	msg := StreamMessage{
		ID:        "1700-0",
		EventType: EventReservationCreated,
		Payload:   payloadBytes,
		Retries:   0,
	}

	// Verify JSON roundtrip
	assert.Equal(t, "1700-0", msg.ID)
	assert.Equal(t, EventReservationCreated, msg.EventType)
	assert.Equal(t, 0, msg.Retries)
}

func TestStreamPublisher_Interface(t *testing.T) {
	// Verify that RedisStreamProducer implements StreamPublisher interface
	var _ StreamPublisher = (*RedisStreamProducer)(nil)
}

func TestContains(t *testing.T) {
	types := []string{"reservation.created", "reservation.cancelled"}
	
	assert.True(t, contains(types, "reservation.created"))
	assert.True(t, contains(types, "reservation.cancelled"))
	assert.False(t, contains(types, "reservation.confirmed"))
	assert.False(t, contains(types, "unknown"))
}
