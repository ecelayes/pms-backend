package entity

import "time"

type Reservation struct {
	ID         string    `json:"id"`
	RoomTypeID string    `json:"room_type_id"`
	GuestEmail string    `json:"guest_email"`
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
	TotalPrice float64   `json:"total_price"`
}

type CreateReservationRequest struct {
	RoomTypeID string `json:"room_type_id"`
	GuestEmail string `json:"guest_email"`
	Start      string `json:"start"` // YYYY-MM-DD
	End        string `json:"end"`   // YYYY-MM-DD
}
