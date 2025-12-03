package entity

import "time"

type Reservation struct {
	BaseEntity
	
	ReservationCode string    `json:"reservation_code"`
	RoomTypeID      string    `json:"room_type_id"`
	GuestEmail      string    `json:"guest_email"`
	Start           time.Time `json:"start"`
	End             time.Time `json:"end"`
	TotalPrice      float64   `json:"total_price"`
	Status          string    `json:"status"`
}

type CreateReservationRequest struct {
	RoomTypeID string `json:"room_type_id"`
	GuestEmail string `json:"guest_email"`
	Start      string `json:"start"`
	End        string `json:"end"`
}
