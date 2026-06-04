package application

import (
	"encoding/json"
	"testing"
	
	"github.com/ecelayes/pms-backend/internal/shared/domain"
)

func TestEventHandlerConstants(t *testing.T) {
	// Verify the event type constants exist and are correct
	if domain.EventReservationCreated != "reservation.created" {
		t.Errorf("Expected EventReservationCreated to be 'reservation.created', got '%s'", domain.EventReservationCreated)
	}
	if domain.EventReservationConfirmed != "reservation.confirmed" {
		t.Errorf("Expected EventReservationConfirmed to be 'reservation.confirmed', got '%s'", domain.EventReservationConfirmed)
	}
	if domain.EventReservationCancelled != "reservation.cancelled" {
		t.Errorf("Expected EventReservationCancelled to be 'reservation.cancelled', got '%s'", domain.EventReservationCancelled)
	}
}

func TestReservationCreatedPayloadJSON(t *testing.T) {
	payload := domain.ReservationCreatedPayload{
		ReservationID:   "res-123",
		PropertyID:      "prop-456",
		UnitTypeID:      "unit-789",
		RatePlanID:      "rate-001",
		GuestID:         "guest-001",
		GuestEmail:      "test@example.com",
		ReservationCode: "ABC12345",
	}
	
	// Test JSON marshaling
	data, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("Failed to marshal ReservationCreatedPayload: %v", err)
	}
	
	// Test JSON unmarshaling
	var decoded domain.ReservationCreatedPayload
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("Failed to unmarshal ReservationCreatedPayload: %v", err)
	}
	
	if decoded.ReservationID != payload.ReservationID {
		t.Errorf("Expected ReservationID '%s', got '%s'", payload.ReservationID, decoded.ReservationID)
	}
	if decoded.GuestEmail != payload.GuestEmail {
		t.Errorf("Expected GuestEmail '%s', got '%s'", payload.GuestEmail, decoded.GuestEmail)
	}
}

func TestReservationConfirmedPayloadJSON(t *testing.T) {
	payload := domain.ReservationConfirmedPayload{
		ReservationID:   "res-123",
		ReservationCode: "ABC12345",
		PropertyID:      "prop-456",
		GuestEmail:      "test@example.com",
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("Failed to marshal ReservationConfirmedPayload: %v", err)
	}
	
	var decoded domain.ReservationConfirmedPayload
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("Failed to unmarshal ReservationConfirmedPayload: %v", err)
	}
	
	if decoded.ReservationCode != payload.ReservationCode {
		t.Errorf("Expected ReservationCode '%s', got '%s'", payload.ReservationCode, decoded.ReservationCode)
	}
}

func TestReservationCancelledPayloadJSON(t *testing.T) {
	payload := domain.ReservationCancelledPayload{
		ReservationID:   "res-123",
		ReservationCode: "ABC12345",
		PropertyID:      "prop-456",
		UnitTypeID:      "unit-789",
		GuestEmail:      "test@example.com",
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("Failed to marshal ReservationCancelledPayload: %v", err)
	}
	
	var decoded domain.ReservationCancelledPayload
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("Failed to unmarshal ReservationCancelledPayload: %v", err)
	}
	
	if decoded.ReservationID != payload.ReservationID {
		t.Errorf("Expected ReservationID '%s', got '%s'", payload.ReservationID, decoded.ReservationID)
	}
}

func TestStreamMessageCreation(t *testing.T) {
	payload := []byte(`{"test":"data"}`)
	msg := domain.StreamMessage{
		ID:        "msg-123",
		EventType: domain.EventReservationCreated,
		Payload:   payload,
		Retries:   0,
	}
	
	if msg.ID != "msg-123" {
		t.Errorf("Expected ID 'msg-123', got '%s'", msg.ID)
	}
	if msg.EventType != domain.EventReservationCreated {
		t.Errorf("Expected EventType 'reservation.created', got '%s'", msg.EventType)
	}
	if string(msg.Payload) != `{"test":"data"}` {
		t.Errorf("Expected Payload '{\"test\":\"data\"}', got '%s'", string(msg.Payload))
	}
}

func TestStreamMessageRetries(t *testing.T) {
	payload := []byte(`{"reservation":"ABC123"}`)
	msg := domain.StreamMessage{
		ID:        "msg-456",
		EventType: domain.EventReservationCancelled,
		Payload:   payload,
		Retries:   2,
	}
	
	if msg.Retries != 2 {
		t.Errorf("Expected Retries 2, got %d", msg.Retries)
	}
}

func TestNewEventHandler(t *testing.T) {
	// Test that NewEventHandler returns a non-nil handler
	// Note: Can't test actual behavior without mock repository
	handler := NewEventHandler(nil)
	if handler == nil {
		t.Error("Expected non-nil EventHandler")
	}
}
