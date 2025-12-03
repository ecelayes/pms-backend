package entity

type RoomType struct {
	BaseEntity
	
	HotelID       string `json:"hotel_id"`
	Name          string `json:"name"`
	Code          string `json:"code"`
	TotalQuantity int    `json:"total_quantity"`
}

type UpdateRoomTypeRequest struct {
	Name          string `json:"name"`
	Code          string `json:"code"`
	TotalQuantity int    `json:"total_quantity"`
}

type AvailabilitySearch struct {
	RoomTypeID   string      `json:"room_type_id"`
	RoomTypeName string      `json:"room_type_name"`
	AvailableQty int         `json:"available_qty"`
	TotalPrice   float64     `json:"total_price"`
	NightlyRates []DailyRate `json:"nightly_rates"`
}

type DailyRate struct {
	Date  string  `json:"date"`
	Price float64 `json:"price"`
}
