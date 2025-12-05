package entity

import "time"

type RoomType struct {
	BaseEntity
	
	HotelID       string   `json:"hotel_id"`
	Name          string   `json:"name"`
	Code          string   `json:"code"`
	TotalQuantity int      `json:"total_quantity"`
	
	MaxOccupancy  int      `json:"max_occupancy"`
	MaxAdults     int      `json:"max_adults"`
	MaxChildren   int      `json:"max_children"`
	
	Amenities     []string `json:"amenities"`
}

type CreateRoomTypeRequest struct {
	HotelID       string   `json:"hotel_id"`
	Name          string   `json:"name"`
	Code          string   `json:"code"`
	TotalQuantity int      `json:"total_quantity"`
	MaxOccupancy  int      `json:"max_occupancy"`
	MaxAdults     int      `json:"max_adults"`
	MaxChildren   int      `json:"max_children"`
	Amenities     []string `json:"amenities"`
}

type UpdateRoomTypeRequest struct {
	Name          string   `json:"name"`
	Code          string   `json:"code"`
	TotalQuantity int      `json:"total_quantity"`
	MaxOccupancy  int      `json:"max_occupancy"`
	MaxAdults     int      `json:"max_adults"`
	MaxChildren   int      `json:"max_children"`
	Amenities     []string `json:"amenities"`
}

type AvailabilityFilter struct {
	Start    time.Time
	End      time.Time
	HotelID  string
	Rooms    int
	Adults   int
	Children int
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

type DailyRate struct {
	Date  string  `json:"date"`
	Price float64 `json:"price"`
}
