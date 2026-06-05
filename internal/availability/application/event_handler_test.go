package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ecelayes/pms-backend/internal/shared/domain"
)

type mockAvailabilityRepoForEventHandler struct {
	updateErr error
}

func (m *mockAvailabilityRepoForEventHandler) UpdateInventory(ctx context.Context, propertyID, unitID string, date time.Time, delta int) error {
	return m.updateErr
}
func (m *mockAvailabilityRepoForEventHandler) GetInventory(ctx context.Context, propertyID, unitID string, date time.Time) (int, error) {
	return 5, nil
}
func (m *mockAvailabilityRepoForEventHandler) GetBatchInventory(ctx context.Context, propertyID, unitID string, dates []time.Time) (map[string]int, error) {
	return map[string]int{}, nil
}

func TestNewEventHandler(t *testing.T) {
	handler := NewEventHandler(nil)
	if handler == nil {
		t.Error("Expected non-nil handler")
	}
}

func TestEventHandler_HandleStreamMessage_ReservationCreated(t *testing.T) {
	repo := &mockAvailabilityRepoForEventHandler{}
	handler := NewEventHandler(repo)

	payload := domain.ReservationCreatedPayload{
		PropertyID:    "prop-1",
		UnitTypeID:    "ut-1",
		ReservationCode: "res-123",
		StartDate:     time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:       time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC),
	}
	payloadBytes, _ := json.Marshal(payload)
	
	msg := &domain.StreamMessage{
		ID:        "msg-123",
		EventType: domain.EventReservationCreated,
		Payload:   payloadBytes,
		Retries:   0,
	}

	err := handler.HandleStreamMessage(msg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestEventHandler_HandleStreamMessage_ReservationConfirmed(t *testing.T) {
	repo := &mockAvailabilityRepoForEventHandler{}
	handler := NewEventHandler(repo)

	payload := domain.ReservationConfirmedPayload{
		ReservationCode: "res-123",
		GuestEmail:       "test@test.com",
	}
	payloadBytes, _ := json.Marshal(payload)
	
	msg := &domain.StreamMessage{
		ID:        "msg-456",
		EventType: domain.EventReservationConfirmed,
		Payload:   payloadBytes,
		Retries:   0,
	}

	err := handler.HandleStreamMessage(msg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestEventHandler_HandleStreamMessage_ReservationCancelled(t *testing.T) {
	repo := &mockAvailabilityRepoForEventHandler{}
	handler := NewEventHandler(repo)

	payload := domain.ReservationCancelledPayload{
		PropertyID:      "prop-1",
		UnitTypeID:      "ut-1",
		ReservationCode: "res-123",
		StartDate:       time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:         time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC),
		GuestEmail:      "test@test.com",
	}
	payloadBytes, _ := json.Marshal(payload)
	
	msg := &domain.StreamMessage{
		ID:        "msg-789",
		EventType: domain.EventReservationCancelled,
		Payload:   payloadBytes,
		Retries:   0,
	}

	err := handler.HandleStreamMessage(msg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestEventHandler_HandleStreamMessage_UnknownEvent(t *testing.T) {
	repo := &mockAvailabilityRepoForEventHandler{}
	handler := NewEventHandler(repo)

	msg := &domain.StreamMessage{
		ID:        "msg-unknown",
		EventType: "unknown.event",
		Payload:   []byte(`{}`),
		Retries:   0,
	}

	err := handler.HandleStreamMessage(msg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestEventHandler_HandleStreamMessage_InvalidPayload(t *testing.T) {
	repo := &mockAvailabilityRepoForEventHandler{}
	handler := NewEventHandler(repo)

	msg := &domain.StreamMessage{
		ID:        "msg-invalid",
		EventType: domain.EventReservationCreated,
		Payload:   []byte(`{invalid json`),
		Retries:   0,
	}

	err := handler.HandleStreamMessage(msg)
	if err == nil {
		t.Error("Expected error for invalid payload")
	}
}

func TestEventHandler_HandleStreamMessage_UpdateInventoryError(t *testing.T) {
	repo := &mockAvailabilityRepoForEventHandler{updateErr: errors.New("update error")}
	handler := NewEventHandler(repo)

	payload := domain.ReservationCreatedPayload{
		PropertyID:      "prop-1",
		UnitTypeID:      "ut-1",
		ReservationCode: "res-123",
		StartDate:       time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:         time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC),
	}
	payloadBytes, _ := json.Marshal(payload)
	
	msg := &domain.StreamMessage{
		ID:        "msg-err",
		EventType: domain.EventReservationCreated,
		Payload:   payloadBytes,
		Retries:   0,
	}

	err := handler.HandleStreamMessage(msg)
	if err == nil {
		t.Error("Expected error from UpdateInventory")
	}
}

func TestEventHandler_HandleReservationCancelled_UpdateError(t *testing.T) {
	repo := &mockAvailabilityRepoForEventHandler{updateErr: errors.New("restore error")}
	handler := NewEventHandler(repo)

	payload := domain.ReservationCancelledPayload{
		PropertyID:      "prop-1",
		UnitTypeID:      "ut-1",
		ReservationCode: "res-123",
		StartDate:       time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:         time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC),
		GuestEmail:      "test@test.com",
	}
	payloadBytes, _ := json.Marshal(payload)
	
	msg := &domain.StreamMessage{
		ID:        "msg-cancel-err",
		EventType: domain.EventReservationCancelled,
		Payload:   payloadBytes,
		Retries:   0,
	}

	err := handler.HandleStreamMessage(msg)
	if err == nil {
		t.Error("Expected error from cancelled reservation handling")
	}
}

func TestEventHandler_WithRepo(t *testing.T) {
	repo := &mockAvailabilityRepoForEventHandler{}
	handler := NewEventHandler(repo)
	
	if handler.repo != repo {
		t.Error("Repo not set correctly")
	}
}

func TestEventHandler_HandleStreamMessage_EmptyPayload(t *testing.T) {
	repo := &mockAvailabilityRepoForEventHandler{}
	handler := NewEventHandler(repo)

	msg := &domain.StreamMessage{
		ID:        "msg-empty",
		EventType: domain.EventReservationConfirmed,
		Payload:   []byte(`{}`),
		Retries:   0,
	}

	err := handler.HandleStreamMessage(msg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestEventHandler_HandleReservationConfirmed_InvalidJSON(t *testing.T) {
	repo := &mockAvailabilityRepoForEventHandler{}
	handler := NewEventHandler(repo)

	msg := &domain.StreamMessage{
		ID:        "msg-confirmed-invalid",
		EventType: domain.EventReservationConfirmed,
		Payload:   []byte(`{invalid json`),
		Retries:   0,
	}

	err := handler.HandleStreamMessage(msg)
	if err == nil {
		t.Error("Expected error for invalid payload in confirmed")
	}
}

func TestEventHandler_HandleReservationCancelled_InvalidJSON(t *testing.T) {
	repo := &mockAvailabilityRepoForEventHandler{}
	handler := NewEventHandler(repo)

	msg := &domain.StreamMessage{
		ID:        "msg-cancelled-invalid",
		EventType: domain.EventReservationCancelled,
		Payload:   []byte(`{invalid json`),
		Retries:   0,
	}

	err := handler.HandleStreamMessage(msg)
	if err == nil {
		t.Error("Expected error for invalid payload in cancelled")
	}
}
