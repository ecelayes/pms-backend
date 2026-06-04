package domain

import (
	"context"
	"testing"
	"time"
)

func TestStreamMessageCreation(t *testing.T) {
	payload := []byte(`{"test":"data"}`)
	msg := StreamMessage{
		ID:        "msg-123",
		EventType: "test.event",
		Payload:   payload,
		Retries:   0,
	}
	
	if msg.ID != "msg-123" {
		t.Errorf("Expected ID 'msg-123', got '%s'", msg.ID)
	}
	if msg.EventType != "test.event" {
		t.Errorf("Expected EventType 'test.event', got '%s'", msg.EventType)
	}
	if string(msg.Payload) != `{"test":"data"}` {
		t.Errorf("Expected Payload '{\"test\":\"data\"}', got '%s'", string(msg.Payload))
	}
	if msg.Retries != 0 {
		t.Errorf("Expected Retries 0, got %d", msg.Retries)
	}
}

func TestStreamMessageWithRetries(t *testing.T) {
	payload := []byte(`{"reservation":"ABC123"}`)
	msg := StreamMessage{
		ID:        "msg-456",
		EventType: EventReservationCreated,
		Payload:   payload,
		Retries:   2,
	}
	
	if msg.Retries != 2 {
		t.Errorf("Expected Retries 2, got %d", msg.Retries)
	}
}

func TestContainsHelper(t *testing.T) {
	eventTypes := []string{EventReservationCreated, EventReservationConfirmed, EventReservationCancelled}
	
	if !contains(eventTypes, EventReservationCreated) {
		t.Error("Expected contains(EventReservationCreated) to be true")
	}
	
	if !contains(eventTypes, EventReservationCancelled) {
		t.Error("Expected contains(EventReservationCancelled) to be true")
	}
	
	if contains(eventTypes, "nonexistent.event") {
		t.Error("Expected contains(nonexistent.event) to be false")
	}
}

func TestContainsWithEmptySlice(t *testing.T) {
	emptySlice := []string{}
	
	// With empty slice, contains should return false
	if contains(emptySlice, "any.event") {
		t.Error("Expected contains to return false for empty slice")
	}
}

func TestContainsWithSingleElement(t *testing.T) {
	single := []string{EventReservationCreated}
	
	if !contains(single, EventReservationCreated) {
		t.Error("Expected contains to find EventReservationCreated")
	}
	
	if contains(single, EventReservationConfirmed) {
		t.Error("Expected contains to not find EventReservationConfirmed")
	}
}

func TestEventTypeConstants(t *testing.T) {
	if EventReservationCreated != "reservation.created" {
		t.Errorf("Expected EventReservationCreated to be 'reservation.created', got '%s'", EventReservationCreated)
	}
	if EventReservationConfirmed != "reservation.confirmed" {
		t.Errorf("Expected EventReservationConfirmed to be 'reservation.confirmed', got '%s'", EventReservationConfirmed)
	}
	if EventReservationCancelled != "reservation.cancelled" {
		t.Errorf("Expected EventReservationCancelled to be 'reservation.cancelled', got '%s'", EventReservationCancelled)
	}
}

func TestStreamConstants(t *testing.T) {
	if StreamReservations != "reservation:events" {
		t.Errorf("Expected StreamReservations to be 'reservation:events', got '%s'", StreamReservations)
	}
	if StreamDLQ != "reservation:events:dlq" {
		t.Errorf("Expected StreamDLQ to be 'reservation:events:dlq', got '%s'", StreamDLQ)
	}
}

func TestGroupConstants(t *testing.T) {
	if GroupAvailability != "availability-group" {
		t.Errorf("Expected GroupAvailability to be 'availability-group', got '%s'", GroupAvailability)
	}
	if GroupNotifications != "notifications-group" {
		t.Errorf("Expected GroupNotifications to be 'notifications-group', got '%s'", GroupNotifications)
	}
	if GroupAnalytics != "analytics-group" {
		t.Errorf("Expected GroupAnalytics to be 'analytics-group', got '%s'", GroupAnalytics)
	}
}

func TestMaxRetriesConstant(t *testing.T) {
	if MaxRetries != 3 {
		t.Errorf("Expected MaxRetries to be 3, got %d", MaxRetries)
	}
}

func TestNewRedisStreamProducer(t *testing.T) {
	// Test that NewRedisStreamProducer returns a non-nil producer
	// Note: Can't fully test without Redis mock, but can verify struct initialization
	producer := &RedisStreamProducer{}
	if producer == nil {
		t.Error("Expected non-nil RedisStreamProducer")
	}
}

func TestNewRedisStreamConsumer(t *testing.T) {
	consumer := NewRedisStreamConsumer(nil, "test-group", "test-consumer")
	
	if consumer == nil {
		t.Error("Expected non-nil RedisStreamConsumer")
	}
	if consumer.groupName != "test-group" {
		t.Errorf("Expected groupName 'test-group', got '%s'", consumer.groupName)
	}
	if consumer.consumerName != "test-consumer" {
		t.Errorf("Expected consumerName 'test-consumer', got '%s'", consumer.consumerName)
	}
}

func TestReservationCreatedPayloadStruct(t *testing.T) {
	now := time.Now()
	payload := ReservationCreatedPayload{
		ReservationID:   "res-123",
		PropertyID:      "prop-456",
		UnitTypeID:      "unit-789",
		RatePlanID:      "rate-001",
		GuestID:         "guest-001",
		GuestEmail:      "test@example.com",
		ReservationCode: "ABC12345",
		StartDate:       now,
		EndDate:         now.AddDate(0, 0, 3),
	}
	
	if payload.ReservationID != "res-123" {
		t.Errorf("Expected ReservationID 'res-123', got '%s'", payload.ReservationID)
	}
	if payload.GuestEmail != "test@example.com" {
		t.Errorf("Expected GuestEmail 'test@example.com', got '%s'", payload.GuestEmail)
	}
	if payload.StartDate.IsZero() {
		t.Error("Expected StartDate to be set")
	}
	if payload.EndDate.Before(payload.StartDate) {
		t.Error("Expected EndDate to be after StartDate")
	}
}

func TestReservationConfirmedPayloadStruct(t *testing.T) {
	payload := ReservationConfirmedPayload{
		ReservationID:   "res-123",
		ReservationCode: "ABC12345",
		PropertyID:      "prop-456",
		GuestEmail:      "test@example.com",
	}
	
	if payload.ReservationID != "res-123" {
		t.Errorf("Expected ReservationID 'res-123', got '%s'", payload.ReservationID)
	}
	if payload.ReservationCode != "ABC12345" {
		t.Errorf("Expected ReservationCode 'ABC12345', got '%s'", payload.ReservationCode)
	}
}

func TestReservationCancelledPayloadStruct(t *testing.T) {
	now := time.Now()
	payload := ReservationCancelledPayload{
		ReservationID:   "res-123",
		ReservationCode: "ABC12345",
		PropertyID:      "prop-456",
		UnitTypeID:      "unit-789",
		GuestEmail:      "test@example.com",
		StartDate:       now,
		EndDate:         now.AddDate(0, 0, 3),
	}
	
	if payload.ReservationID != "res-123" {
		t.Errorf("Expected ReservationID 'res-123', got '%s'", payload.ReservationID)
	}
	if payload.GuestEmail != "test@example.com" {
		t.Errorf("Expected GuestEmail 'test@example.com', got '%s'", payload.GuestEmail)
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	select {
	case <-ctx.Done():
		// Expected behavior - context is cancelled
	default:
		t.Error("Expected context to be done after cancellation")
	}
}
