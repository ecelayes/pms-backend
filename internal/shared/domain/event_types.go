package domain

import (
	"time"
)

const (
	EventReservationCreated   = "reservation.created"
	EventReservationConfirmed = "reservation.confirmed"
	EventReservationCancelled = "reservation.cancelled"
)

type ReservationCreatedPayload struct {
	ReservationID   string    `json:"reservation_id"`
	PropertyID      string    `json:"property_id"`
	UnitTypeID      string    `json:"unit_type_id"`
	RatePlanID      string    `json:"rate_plan_id"`
	GuestID         string    `json:"guest_id"`
	GuestEmail      string    `json:"guest_email"`
	ReservationCode string    `json:"reservation_code"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
}
type ReservationConfirmedPayload struct {
	ReservationID   string `json:"reservation_id"`
	ReservationCode string `json:"reservation_code"`
	PropertyID      string `json:"property_id"`
	GuestEmail      string `json:"guest_email"`
}
type ReservationCancelledPayload struct {
	ReservationID   string    `json:"reservation_id"`
	ReservationCode string    `json:"reservation_code"`
	PropertyID      string    `json:"property_id"`
	UnitTypeID      string    `json:"unit_type_id"`
	GuestEmail      string    `json:"guest_email"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
}
