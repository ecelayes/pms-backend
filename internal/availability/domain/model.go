package domain

import (
	"time"
)

type AvailabilityReadModel struct {
	PropertyID string    `json:"property_id"`
	UnitID     string    `json:"unit_id"`
	Date       time.Time `json:"date"`
	Count      int       `json:"count"`
}
type SearchCriteria struct {
	PropertyID string
	StartDate  time.Time
	EndDate    time.Time
}
