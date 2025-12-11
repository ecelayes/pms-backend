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

type RateOption struct {
	RatePlanID          string             `json:"rate_plan_id"`
	RatePlanName        string             `json:"rate_plan_name"`
	Description         string             `json:"description"`
	TotalPrice          float64            `json:"total_price"`
	CancellationPolicy  CancellationPolicy `json:"cancellation_policy"`
	MealPlan            MealPlan           `json:"meal_plan"`
	PaymentPolicy       PaymentPolicy      `json:"payment_policy"`
	NightlyRates        []DailyRate        `json:"nightly_rates"`
}

type AvailabilitySearch struct {
	RoomTypeID   string      `json:"room_type_id"`
	RoomTypeName string      `json:"room_type_name"`
	AvailableQty int         `json:"available_qty"`
	MaxOccupancy int         `json:"max_occupancy"`
	MaxAdults    int         `json:"max_adults"`
	MaxChildren  int         `json:"max_children"`
	
	Amenities    []string    `json:"amenities"`

	Rates        []RateOption `json:"rates"`
}
