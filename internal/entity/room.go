package entity

type RoomType struct {
	BaseEntity
	
	HotelID       string   `json:"hotel_id"`
	Name          string   `json:"name"`
	Code          string   `json:"code"`
	TotalQuantity int      `json:"total_quantity"`

	BasePrice     float64  `json:"base_price"`
	
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
	BasePrice     float64  `json:"base_price"`
	MaxOccupancy  int      `json:"max_occupancy"`
	MaxAdults     int      `json:"max_adults"`
	MaxChildren   int      `json:"max_children"`
	Amenities     []string `json:"amenities"`
}

type UpdateRoomTypeRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`

	TotalQuantity *int     `json:"total_quantity"`
	MaxOccupancy  *int     `json:"max_occupancy"`
	MaxAdults     *int     `json:"max_adults"`
	MaxChildren   *int     `json:"max_children"`
	
	BasePrice     *float64 `json:"base_price"`

	Amenities []string `json:"amenities"`
}
