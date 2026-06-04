package domain

import (
	"testing"
	"time"

	"github.com/ecelayes/pms-backend/internal/shared/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
)

func TestReservationCreatedEventName(t *testing.T) {
	event := &ReservationCreatedEvent{
		OccurredAt: time.Now(),
	}
	
	if event.Name() != domain.EventReservationCreated {
		t.Errorf("Expected Name '%s', got '%s'", domain.EventReservationCreated, event.Name())
	}
}

func TestReservationCreatedEventOccurredOn(t *testing.T) {
	now := time.Now()
	event := &ReservationCreatedEvent{
		OccurredAt: now,
	}
	
	if !event.OccurredOn().Equal(now) {
		t.Errorf("Expected OccurredOn %v, got %v", now, event.OccurredOn())
	}
}

func TestReservationCreatedEventToPayload(t *testing.T) {
	event := &ReservationCreatedEvent{
		OccurredAt:      time.Now(),
		ReservationID:   "res-123",
		PropertyID:      "prop-456",
		UnitTypeID:      "ut-789",
		RatePlanID:      "rp-001",
		GuestID:         "guest-001",
		GuestEmail:      "test@example.com",
		DateRange:       mustDateRange(t, "2024-06-01", "2024-06-05"),
		Price:           mustMoney(t, 50000, "USD"),
		ReservationCode: "ABC12345",
	}
	
	payload := event.ToPayload()
	
	if payload.ReservationID != "res-123" {
		t.Errorf("Expected ReservationID 'res-123', got '%s'", payload.ReservationID)
	}
	if payload.PropertyID != "prop-456" {
		t.Errorf("Expected PropertyID 'prop-456', got '%s'", payload.PropertyID)
	}
	if payload.UnitTypeID != "ut-789" {
		t.Errorf("Expected UnitTypeID 'ut-789', got '%s'", payload.UnitTypeID)
	}
	if payload.GuestEmail != "test@example.com" {
		t.Errorf("Expected GuestEmail 'test@example.com', got '%s'", payload.GuestEmail)
	}
	if payload.ReservationCode != "ABC12345" {
		t.Errorf("Expected ReservationCode 'ABC12345', got '%s'", payload.ReservationCode)
	}
}

func TestReservationConfirmedEventName(t *testing.T) {
	event := &ReservationConfirmedEvent{
		OccurredAt: time.Now(),
	}
	
	if event.Name() != domain.EventReservationConfirmed {
		t.Errorf("Expected Name '%s', got '%s'", domain.EventReservationConfirmed, event.Name())
	}
}

func TestReservationConfirmedEventOccurredOn(t *testing.T) {
	now := time.Now()
	event := &ReservationConfirmedEvent{
		OccurredAt: now,
	}
	
	if !event.OccurredOn().Equal(now) {
		t.Errorf("Expected OccurredOn %v, got %v", now, event.OccurredOn())
	}
}

func TestReservationConfirmedEventToPayload(t *testing.T) {
	event := &ReservationConfirmedEvent{
		OccurredAt:      time.Now(),
		ReservationID:   "res-456",
		ReservationCode: "XYZ98765",
		PropertyID:      "prop-789",
		GuestEmail:      "guest@example.com",
		DateRange:       mustDateRange(t, "2024-07-01", "2024-07-03"),
	}
	
	payload := event.ToPayload()
	
	if payload.ReservationID != "res-456" {
		t.Errorf("Expected ReservationID 'res-456', got '%s'", payload.ReservationID)
	}
	if payload.ReservationCode != "XYZ98765" {
		t.Errorf("Expected ReservationCode 'XYZ98765', got '%s'", payload.ReservationCode)
	}
	if payload.PropertyID != "prop-789" {
		t.Errorf("Expected PropertyID 'prop-789', got '%s'", payload.PropertyID)
	}
	if payload.GuestEmail != "guest@example.com" {
		t.Errorf("Expected GuestEmail 'guest@example.com', got '%s'", payload.GuestEmail)
	}
}

func TestReservationCancelledEventName(t *testing.T) {
	event := &ReservationCancelledEvent{
		OccurredAt: time.Now(),
	}
	
	if event.Name() != domain.EventReservationCancelled {
		t.Errorf("Expected Name '%s', got '%s'", domain.EventReservationCancelled, event.Name())
	}
}

func TestReservationCancelledEventOccurredOn(t *testing.T) {
	now := time.Now()
	event := &ReservationCancelledEvent{
		OccurredAt: now,
	}
	
	if !event.OccurredOn().Equal(now) {
		t.Errorf("Expected OccurredOn %v, got %v", now, event.OccurredOn())
	}
}

func TestReservationCancelledEventToPayload(t *testing.T) {
	event := &ReservationCancelledEvent{
		OccurredAt:      time.Now(),
		ReservationID:   "res-789",
		ReservationCode: "CANCEL123",
		PropertyID:      "prop-abc",
		UnitTypeID:      "ut-xyz",
		GuestEmail:      "cancelled@example.com",
		DateRange:       mustDateRange(t, "2024-08-01", "2024-08-10"),
	}
	
	payload := event.ToPayload()
	
	if payload.ReservationID != "res-789" {
		t.Errorf("Expected ReservationID 'res-789', got '%s'", payload.ReservationID)
	}
	if payload.ReservationCode != "CANCEL123" {
		t.Errorf("Expected ReservationCode 'CANCEL123', got '%s'", payload.ReservationCode)
	}
	if payload.PropertyID != "prop-abc" {
		t.Errorf("Expected PropertyID 'prop-abc', got '%s'", payload.PropertyID)
	}
	if payload.UnitTypeID != "ut-xyz" {
		t.Errorf("Expected UnitTypeID 'ut-xyz', got '%s'", payload.UnitTypeID)
	}
	if payload.GuestEmail != "cancelled@example.com" {
		t.Errorf("Expected GuestEmail 'cancelled@example.com', got '%s'", payload.GuestEmail)
	}
}

func mustDateRange(t *testing.T, startStr, endStr string) vo.DateRange {
	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)
	dr, err := vo.NewDateRange(start, end)
	if err != nil {
		t.Fatalf("Failed to create date range: %v", err)
	}
	return dr
}

func mustMoney(t *testing.T, amount int64, currency string) vo.Money {
	return vo.NewMoney(amount, currency)
}
