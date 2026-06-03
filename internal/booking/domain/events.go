package domain

import (
	sharedDomain "github.com/ecelayes/pms-backend/internal/shared/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"time"
)

type ReservationCreatedEvent struct {
	OccurredAt      time.Time
	ReservationID   string
	PropertyID      string
	UnitTypeID      string
	RatePlanID      string
	GuestID         string
	GuestEmail      string
	DateRange       vo.DateRange
	Price           vo.Money
	ReservationCode string
}

func (e *ReservationCreatedEvent) Name() string {
	return sharedDomain.EventReservationCreated
}
func (e *ReservationCreatedEvent) OccurredOn() time.Time {
	return e.OccurredAt
}
func NewReservationCreatedEvent(r *Reservation) *ReservationCreatedEvent {
	return &ReservationCreatedEvent{
		OccurredAt:      time.Now(),
		ReservationID:   r.ID(),
		PropertyID:      r.PropertyID(),
		UnitTypeID:      r.UnitTypeID(),
		RatePlanID:      r.RatePlanID(),
		GuestID:         r.GuestID(),
		GuestEmail:      r.GuestEmail(),
		DateRange:       r.DateRange(),
		Price:           r.Price(),
		ReservationCode: r.ReservationCode(),
	}
}
func (e *ReservationCreatedEvent) ToPayload() sharedDomain.ReservationCreatedPayload {
	return sharedDomain.ReservationCreatedPayload{
		ReservationID:   e.ReservationID,
		PropertyID:      e.PropertyID,
		UnitTypeID:      e.UnitTypeID,
		RatePlanID:      e.RatePlanID,
		GuestID:         e.GuestID,
		GuestEmail:      e.GuestEmail,
		ReservationCode: e.ReservationCode,
		StartDate:       e.DateRange.Start(),
		EndDate:         e.DateRange.End(),
	}
}

type ReservationConfirmedEvent struct {
	OccurredAt      time.Time
	ReservationID   string
	ReservationCode string
	PropertyID      string
	GuestEmail      string
	DateRange       vo.DateRange
}

func (e *ReservationConfirmedEvent) Name() string {
	return sharedDomain.EventReservationConfirmed
}
func (e *ReservationConfirmedEvent) OccurredOn() time.Time {
	return e.OccurredAt
}
func NewReservationConfirmedEvent(r *Reservation) *ReservationConfirmedEvent {
	return &ReservationConfirmedEvent{
		OccurredAt:      time.Now(),
		ReservationID:   r.ID(),
		ReservationCode: r.ReservationCode(),
		PropertyID:      r.PropertyID(),
		GuestEmail:      r.GuestEmail(),
		DateRange:       r.DateRange(),
	}
}
func (e *ReservationConfirmedEvent) ToPayload() sharedDomain.ReservationConfirmedPayload {
	return sharedDomain.ReservationConfirmedPayload{
		ReservationID:   e.ReservationID,
		ReservationCode: e.ReservationCode,
		PropertyID:      e.PropertyID,
		GuestEmail:      e.GuestEmail,
	}
}

type ReservationCancelledEvent struct {
	OccurredAt      time.Time
	ReservationID   string
	ReservationCode string
	PropertyID      string
	UnitTypeID      string
	GuestEmail      string
	DateRange       vo.DateRange
}

func (e *ReservationCancelledEvent) Name() string {
	return sharedDomain.EventReservationCancelled
}
func (e *ReservationCancelledEvent) OccurredOn() time.Time {
	return e.OccurredAt
}
func NewReservationCancelledEvent(r *Reservation) *ReservationCancelledEvent {
	return &ReservationCancelledEvent{
		OccurredAt:      time.Now(),
		ReservationID:   r.ID(),
		ReservationCode: r.ReservationCode(),
		PropertyID:      r.PropertyID(),
		UnitTypeID:      r.UnitTypeID(),
		GuestEmail:      r.GuestEmail(),
		DateRange:       r.DateRange(),
	}
}
func (e *ReservationCancelledEvent) ToPayload() sharedDomain.ReservationCancelledPayload {
	return sharedDomain.ReservationCancelledPayload{
		ReservationID:   e.ReservationID,
		ReservationCode: e.ReservationCode,
		PropertyID:      e.PropertyID,
		UnitTypeID:      e.UnitTypeID,
		GuestEmail:      e.GuestEmail,
		StartDate:       e.DateRange.Start(),
		EndDate:         e.DateRange.End(),
	}
}
