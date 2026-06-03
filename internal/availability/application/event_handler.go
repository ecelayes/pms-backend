package application

import (
	"context"
	"encoding/json"
	availDomain "github.com/ecelayes/pms-backend/internal/availability/domain"
	"github.com/ecelayes/pms-backend/internal/shared/domain"
	"log"
)

type EventHandler struct {
	repo availDomain.AvailabilityRepository
}

func NewEventHandler(repo availDomain.AvailabilityRepository) *EventHandler {
	return &EventHandler{repo: repo}
}
func (h *EventHandler) HandleStreamMessage(msg *domain.StreamMessage) error {
	log.Printf("[AvailabilityEventHandler] Processing stream message %s of type %s", msg.ID, msg.EventType)
	switch msg.EventType {
	case domain.EventReservationCreated:
		return h.handleReservationCreated(msg)
	case domain.EventReservationConfirmed:
		return h.handleReservationConfirmed(msg)
	case domain.EventReservationCancelled:
		return h.handleReservationCancelled(msg)
	default:
		log.Printf("[AvailabilityEventHandler] Unknown event type: %s", msg.EventType)
		return nil
	}
}
func (h *EventHandler) handleReservationCreated(msg *domain.StreamMessage) error {
	var payload domain.ReservationCreatedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}
	log.Printf("[AvailabilityEventHandler] Processing reservation.created for unit %s from %v to %v",
		payload.UnitTypeID, payload.StartDate, payload.EndDate)
	delta := -1
	ctx := context.Background()
	for d := payload.StartDate; d.Before(payload.EndDate); d = d.AddDate(0, 0, 1) {
		if err := h.repo.UpdateInventory(ctx, payload.PropertyID, payload.UnitTypeID, d, delta); err != nil {
			log.Printf("[AvailabilityEventHandler] Failed to update inventory: %v", err)
			return err
		}
	}
	log.Printf("[AvailabilityEventHandler] Successfully decremented inventory for reservation %s", payload.ReservationCode)
	return nil
}
func (h *EventHandler) handleReservationConfirmed(msg *domain.StreamMessage) error {
	var payload domain.ReservationConfirmedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}
	log.Printf("[AvailabilityEventHandler] Processing reservation.confirmed for reservation %s", payload.ReservationCode)
	return nil
}
func (h *EventHandler) handleReservationCancelled(msg *domain.StreamMessage) error {
	var payload domain.ReservationCancelledPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}
	log.Printf("[AvailabilityEventHandler] Processing reservation.cancelled for reservation %s from %v to %v",
		payload.ReservationCode, payload.StartDate, payload.EndDate)
	delta := 1
	ctx := context.Background()
	for d := payload.StartDate; d.Before(payload.EndDate); d = d.AddDate(0, 0, 1) {
		if err := h.repo.UpdateInventory(ctx, payload.PropertyID, payload.UnitTypeID, d, delta); err != nil {
			log.Printf("[AvailabilityEventHandler] Failed to restore inventory on cancellation: %v", err)
			return err
		}
	}
	log.Printf("[AvailabilityEventHandler] Successfully restored inventory for cancelled reservation %s", payload.ReservationCode)
	return nil
}
