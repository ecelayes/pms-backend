package entity

import "time"

type AvailabilityFilter struct {
	HotelID  string    `json:"hotel_id"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Adults   int       `json:"adults"`
	Children int       `json:"children"`
	Rooms    int       `json:"rooms"`
}

type DailyRate struct {
	Date  string  `json:"date"`
	Price float64 `json:"price"`
}

type AvailabilitySearch struct {
	RoomTypeID   string      `json:"room_type_id"`
	RoomTypeName string      `json:"room_type_name"`
	AvailableQty int         `json:"available_qty"`
	TotalPrice   float64     `json:"total_price"`
	MaxOccupancy int         `json:"max_occupancy"`
	MaxAdults    int         `json:"max_adults"`
	MaxChildren  int         `json:"max_children"`
	Amenities    []string    `json:"amenities"`
	NightlyRates []DailyRate `json:"nightly_rates"`
}
