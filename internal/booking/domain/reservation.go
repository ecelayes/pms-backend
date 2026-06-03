package domain

import (
	"errors"
	"github.com/ecelayes/pms-backend/internal/shared/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"github.com/google/uuid"
	"time"
)

var (
	ErrInvalidReservationDates     = errors.New("reservation dates must be in the future")
	ErrReservationAlreadyConfirmed = errors.New("reservation is already confirmed")
)

type ReservationStatus string

const (
	StatusPending   ReservationStatus = "pending"
	StatusConfirmed ReservationStatus = "confirmed"
	StatusCancelled ReservationStatus = "cancelled"
)

type Reservation struct {
	id              string
	propertyID      string
	unitID          string
	unitTypeID      string
	ratePlanID      string
	guestID         string
	dateRange       vo.DateRange
	price           vo.Money
	status          ReservationStatus
	createdAt       time.Time
	guestEmail      string
	reservationCode string
	pendingEvents   []domain.DomainEvent
}

func NewReservation(propertyID, unitTypeID, ratePlanID, guestID string, dateRange vo.DateRange, price vo.Money, guestEmail string) (*Reservation, error) {
	if dateRange.Start().Before(time.Now()) {
		return nil, ErrInvalidReservationDates
	}
	res := &Reservation{
		id:              uuid.New().String(),
		propertyID:      propertyID,
		unitTypeID:      unitTypeID,
		ratePlanID:      ratePlanID,
		guestID:         guestID,
		dateRange:       dateRange,
		price:           price,
		status:          StatusPending,
		createdAt:       time.Now(),
		guestEmail:      guestEmail,
		reservationCode: uuid.New().String()[:8],
		pendingEvents:   []domain.DomainEvent{},
	}
	res.pendingEvents = append(res.pendingEvents, NewReservationCreatedEvent(res))
	return res, nil
}
func Reconstitute(
	id, propertyID, unitTypeID, ratePlanID, unitID, guestID string,
	dateRange vo.DateRange,
	price vo.Money,
	status ReservationStatus,
	guestEmail, reservationCode string,
	createdAt time.Time,
) *Reservation {
	return &Reservation{
		id:              id,
		propertyID:      propertyID,
		unitTypeID:      unitTypeID,
		ratePlanID:      ratePlanID,
		unitID:          unitID,
		guestID:         guestID,
		dateRange:       dateRange,
		price:           price,
		status:          status,
		guestEmail:      guestEmail,
		reservationCode: reservationCode,
		createdAt:       createdAt,
		pendingEvents:   []domain.DomainEvent{},
	}
}
func (r *Reservation) Confirm() error {
	if r.status == StatusConfirmed {
		return ErrReservationAlreadyConfirmed
	}
	r.status = StatusConfirmed
	r.pendingEvents = append(r.pendingEvents, NewReservationConfirmedEvent(r))
	return nil
}
func (r *Reservation) Cancel() {
	r.status = StatusCancelled
	r.pendingEvents = append(r.pendingEvents, NewReservationCancelledEvent(r))
}
func (r *Reservation) PopEvents() []domain.DomainEvent {
	events := r.pendingEvents
	r.pendingEvents = []domain.DomainEvent{}
	return events
}
func (r *Reservation) ID() string {
	return r.id
}
func (r *Reservation) PropertyID() string {
	return r.propertyID
}
func (r *Reservation) UnitID() string {
	return r.unitID
}
func (r *Reservation) UnitTypeID() string {
	return r.unitTypeID
}
func (r *Reservation) RatePlanID() string {
	return r.ratePlanID
}
func (r *Reservation) GuestID() string {
	return r.guestID
}
func (r *Reservation) DateRange() vo.DateRange {
	return r.dateRange
}
func (r *Reservation) Price() vo.Money {
	return r.price
}
func (r *Reservation) Status() ReservationStatus {
	return r.status
}
func (r *Reservation) GuestEmail() string {
	return r.guestEmail
}
func (r *Reservation) ReservationCode() string {
	return r.reservationCode
}
