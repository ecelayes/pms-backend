package entity

import "time"

type PriceRule struct {
	BaseEntity
	
	UnitTypeID string    `json:"unit_type_id"`
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
	Price      float64   `json:"price"`
}

type SetPriceRequest struct {
	UnitTypeID string  `json:"unit_type_id"`
	Start      string  `json:"start"`
	End        string  `json:"end"`
	Price      float64 `json:"price"`
}
