package entity

import "time"

type Hotel struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateHotelRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}
