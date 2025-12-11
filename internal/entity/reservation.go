package entity

import "time"

type Reservation struct {
	BaseEntity
	
	ReservationCode string    `json:"reservation_code"`
	RoomTypeID      string    `json:"room_type_id"`
	RatePlanID      *string   `json:"rate_plan_id,omitempty"`
	GuestID         string    `json:"guest_id"`
	Start           time.Time `json:"start"`
	End             time.Time `json:"end"`
	TotalPrice      float64   `json:"total_price"`
	Status          string    `json:"status"`
	
	Adults          int       `json:"adults"`
	Children        int       `json:"children"`
}

type CreateReservationRequest struct {
	RoomTypeID string  `json:"room_type_id"`
	RatePlanID *string `json:"rate_plan_id"`
	
	GuestEmail     string `json:"guest_email"`
	GuestFirstName string `json:"guest_first_name"`
	GuestLastName  string `json:"guest_last_name"`
	GuestPhone     string `json:"guest_phone"`
	
	Start    string `json:"start"`
	End      string `json:"end"`
	
	Adults   int `json:"adults"`
	Children int `json:"children"`
}
