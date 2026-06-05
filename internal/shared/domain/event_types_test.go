package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEventReservationCreated_Constants(t *testing.T) {
	if EventReservationCreated != "reservation.created" {
		t.Errorf("Expected 'reservation.created', got '%s'", EventReservationCreated)
	}
	if EventReservationConfirmed != "reservation.confirmed" {
		t.Errorf("Expected 'reservation.confirmed', got '%s'", EventReservationConfirmed)
	}
	if EventReservationCancelled != "reservation.cancelled" {
		t.Errorf("Expected 'reservation.cancelled', got '%s'", EventReservationCancelled)
	}
}

func TestReservationCreatedPayload_JSON(t *testing.T) {
	payload := ReservationCreatedPayload{
		ReservationID:   "res-123",
		PropertyID:       "prop-1",
		UnitTypeID:       "ut-1",
		RatePlanID:       "rp-1",
		GuestID:          "guest-1",
		GuestEmail:       "test@test.com",
		ReservationCode:  "ABC123",
		StartDate:        time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:          time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("Failed to marshal: %v", err)
	}

	var parsed ReservationCreatedPayload
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}

	if parsed.ReservationID != "res-123" {
		t.Errorf("Expected ReservationID 'res-123', got '%s'", parsed.ReservationID)
	}
	if parsed.PropertyID != "prop-1" {
		t.Errorf("Expected PropertyID 'prop-1', got '%s'", parsed.PropertyID)
	}
	if parsed.UnitTypeID != "ut-1" {
		t.Errorf("Expected UnitTypeID 'ut-1', got '%s'", parsed.UnitTypeID)
	}
	if parsed.GuestEmail != "test@test.com" {
		t.Errorf("Expected GuestEmail 'test@test.com', got '%s'", parsed.GuestEmail)
	}
	if parsed.ReservationCode != "ABC123" {
		t.Errorf("Expected ReservationCode 'ABC123', got '%s'", parsed.ReservationCode)
	}
}

func TestReservationConfirmedPayload_JSON(t *testing.T) {
	payload := ReservationConfirmedPayload{
		ReservationID:   "res-456",
		ReservationCode:  "DEF456",
		PropertyID:       "prop-2",
		GuestEmail:       "user@example.com",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("Failed to marshal: %v", err)
	}

	var parsed ReservationConfirmedPayload
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}

	if parsed.ReservationID != "res-456" {
		t.Errorf("Expected ReservationID 'res-456', got '%s'", parsed.ReservationID)
	}
	if parsed.ReservationCode != "DEF456" {
		t.Errorf("Expected ReservationCode 'DEF456', got '%s'", parsed.ReservationCode)
	}
	if parsed.GuestEmail != "user@example.com" {
		t.Errorf("Expected GuestEmail 'user@example.com', got '%s'", parsed.GuestEmail)
	}
}

func TestReservationCancelledPayload_JSON(t *testing.T) {
	payload := ReservationCancelledPayload{
		ReservationID:   "res-789",
		ReservationCode: "GHI789",
		PropertyID:      "prop-3",
		UnitTypeID:      "ut-3",
		GuestEmail:      "cancel@example.com",
		StartDate:       time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
		EndDate:         time.Date(2024, 7, 3, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("Failed to marshal: %v", err)
	}

	var parsed ReservationCancelledPayload
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}

	if parsed.ReservationID != "res-789" {
		t.Errorf("Expected ReservationID 'res-789', got '%s'", parsed.ReservationID)
	}
	if parsed.UnitTypeID != "ut-3" {
		t.Errorf("Expected UnitTypeID 'ut-3', got '%s'", parsed.UnitTypeID)
	}
	if parsed.GuestEmail != "cancel@example.com" {
		t.Errorf("Expected GuestEmail 'cancel@example.com', got '%s'", parsed.GuestEmail)
	}
}

func TestReservationCreatedPayload_EmptyFields(t *testing.T) {
	jsonStr := `{"reservation_id":"","property_id":"","unit_type_id":""}`
	var payload ReservationCreatedPayload
	err := json.Unmarshal([]byte(jsonStr), &payload)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}
	if payload.ReservationID != "" {
		t.Errorf("Expected empty ReservationID, got '%s'", payload.ReservationID)
	}
}

func TestReservationConfirmedPayload_EmptyFields(t *testing.T) {
	jsonStr := `{"reservation_id":"","reservation_code":""}`
	var payload ReservationConfirmedPayload
	err := json.Unmarshal([]byte(jsonStr), &payload)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}
	if payload.ReservationID != "" {
		t.Errorf("Expected empty ReservationID, got '%s'", payload.ReservationID)
	}
}

func TestReservationCancelledPayload_EmptyFields(t *testing.T) {
	jsonStr := `{"reservation_id":"","property_id":"","unit_type_id":""}`
	var payload ReservationCancelledPayload
	err := json.Unmarshal([]byte(jsonStr), &payload)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}
	if payload.ReservationID != "" {
		t.Errorf("Expected empty ReservationID, got '%s'", payload.ReservationID)
	}
}
