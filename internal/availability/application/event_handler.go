package application

import (
	"context"
	"encoding/json"

	availDomain "github.com/ecelayes/pms-backend/internal/availability/domain"
	"github.com/ecelayes/pms-backend/internal/shared/domain"
	"go.uber.org/zap"
)

type EventHandler struct {
	repo   availDomain.AvailabilityRepository
	logger *zap.Logger
}

func NewEventHandler(repo availDomain.AvailabilityRepository, logger *zap.Logger) *EventHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &EventHandler{repo: repo, logger: logger}
}

func (h *EventHandler) HandleStreamMessage(msg *domain.StreamMessage) error {
	h.logger.Debug("processing stream message",
		zap.String("message_id", msg.ID),
		zap.String("event_type", msg.EventType))
	switch msg.EventType {
	case domain.EventReservationCreated:
		return h.handleReservationCreated(msg)
	case domain.EventReservationConfirmed:
		return h.handleReservationConfirmed(msg)
	case domain.EventReservationCancelled:
		return h.handleReservationCancelled(msg)
	default:
		h.logger.Warn("unknown event type", zap.String("event_type", msg.EventType))
		return nil
	}
}

func (h *EventHandler) handleReservationCreated(msg *domain.StreamMessage) error {
	var payload domain.ReservationCreatedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.logger.Error("failed to unmarshal reservation.created payload", zap.Error(err))
		return err
	}
	h.logger.Info("processing reservation.created",
		zap.String("unit_type_id", payload.UnitTypeID),
		zap.Time("start_date", payload.StartDate),
		zap.Time("end_date", payload.EndDate))
	delta := -1
	ctx := context.Background()
	for d := payload.StartDate; d.Before(payload.EndDate); d = d.AddDate(0, 0, 1) {
		if err := h.repo.UpdateInventory(ctx, payload.PropertyID, payload.UnitTypeID, d, delta); err != nil {
			h.logger.Error("failed to update inventory",
				zap.String("property_id", payload.PropertyID),
				zap.String("unit_type_id", payload.UnitTypeID),
				zap.Time("date", d),
				zap.Error(err))
			return err
		}
	}
	h.logger.Info("decremented inventory",
		zap.String("reservation_code", payload.ReservationCode))
	return nil
}

func (h *EventHandler) handleReservationConfirmed(msg *domain.StreamMessage) error {
	var payload domain.ReservationConfirmedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.logger.Error("failed to unmarshal reservation.confirmed payload", zap.Error(err))
		return err
	}
	h.logger.Info("processing reservation.confirmed",
		zap.String("reservation_code", payload.ReservationCode))
	return nil
}

func (h *EventHandler) handleReservationCancelled(msg *domain.StreamMessage) error {
	var payload domain.ReservationCancelledPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.logger.Error("failed to unmarshal reservation.cancelled payload", zap.Error(err))
		return err
	}
	h.logger.Info("processing reservation.cancelled",
		zap.String("reservation_code", payload.ReservationCode),
		zap.Time("start_date", payload.StartDate),
		zap.Time("end_date", payload.EndDate))
	delta := 1
	ctx := context.Background()
	for d := payload.StartDate; d.Before(payload.EndDate); d = d.AddDate(0, 0, 1) {
		if err := h.repo.UpdateInventory(ctx, payload.PropertyID, payload.UnitTypeID, d, delta); err != nil {
			h.logger.Error("failed to restore inventory on cancellation",
				zap.String("property_id", payload.PropertyID),
				zap.String("unit_type_id", payload.UnitTypeID),
				zap.Time("date", d),
				zap.Error(err))
			return err
		}
	}
	h.logger.Info("restored inventory",
		zap.String("reservation_code", payload.ReservationCode))
	return nil
}
