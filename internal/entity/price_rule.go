package entity

import "time"

type PriceRule struct {
	BaseEntity
	
	RoomTypeID string    `json:"room_type_id"`
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
	Price      float64   `json:"price"`
	Priority   int       `json:"priority"`
}

type CreatePriceRuleRequest struct {
	RoomTypeID string  `json:"room_type_id"`
	Start      string  `json:"start"`
	End        string  `json:"end"`
	Price      float64 `json:"price"`
	Priority   int     `json:"priority"`
}

type UpdatePriceRuleRequest struct {
	Price    float64 `json:"price"`
	Priority int     `json:"priority"`
}
