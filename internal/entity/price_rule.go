package entity

import "time"

type PriceRule struct {
	ID         string    `json:"id"`
	RoomTypeID string    `json:"room_type_id"`
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
	Price      float64   `json:"price"`
	Priority   int       `json:"priority"`
}

type CreatePriceRuleRequest struct {
	RoomTypeID string  `json:"room_type_id"`
	Start      string  `json:"start"` // YYYY-MM-DD
	End        string  `json:"end"`   // YYYY-MM-DD
	Price      float64 `json:"price"`
	Priority   int     `json:"priority"`
}
