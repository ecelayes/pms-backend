package entity

import "time"

type Reservation struct {
	ID              string    `json:"id"`
	ReservationCode string    `json:"reservation_code"`
	RoomTypeID      string    `json:"room_type_id"`
	GuestEmail      string    `json:"guest_email"`
	Start           time.Time `json:"start"`	// YYYY-MM-DD
	End             time.Time `json:"end"`		// YYYY-MM-DD
	TotalPrice      float64   `json:"total_price"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateReservationRequest struct {
	RoomTypeID string `json:"room_type_id"`
	GuestEmail string `json:"guest_email"`
	Start      string `json:"start"`	// YYYY-MM-DD
	End        string `json:"end"`		// YYYY-MM-DD
}
