package dto

type PaginatedResponse[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}
type Meta struct {
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
}
type AvailabilityResult struct {
	UnitTypeID   string  `json:"unit_type_id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	AvailableQty int     `json:"available_quantity"`
	TotalPrice   float64 `json:"total_price"`
	Currency     string  `json:"currency"`
	BasePrice    float64 `json:"base_price"`
}
type Amenity struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}
type GuestService struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}
type Property struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	Type           string `json:"type"`
}
type UnitType struct {
	ID           string   `json:"id"`
	PropertyID   string   `json:"property_id"`
	Name         string   `json:"name"`
	Code         string   `json:"code"`
	TotalQty     int      `json:"total_quantity"`
	BasePrice    float64  `json:"base_price"`
	MaxOccupancy int      `json:"max_occupancy"`
	Amenities    []string `json:"amenities"`
}
type Unit struct {
	ID         string `json:"id"`
	PropertyID string `json:"property_id"`
	UnitTypeID string `json:"unit_type_id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
}
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}
type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}
type RatePlan struct {
	ID         string `json:"id"`
	PropertyID string `json:"property_id"`
	UnitTypeID string `json:"unit_type_id"`
	Name       string `json:"name"`
}
type PriceRule struct {
	ID         string  `json:"id"`
	UnitTypeID string  `json:"unit_type_id"`
	Start      string  `json:"start"`
	End        string  `json:"end"`
	Price      float64 `json:"price"`
	Currency   string  `json:"currency"`
}
type Reservation struct {
	ID              string  `json:"id"`
	ReservationCode string  `json:"reservation_code"`
	GuestEmail      string  `json:"guest_email"`
	Status          string  `json:"status"`
	UnitTypeID      string  `json:"unit_type_id"`
	Start           string  `json:"start"`
	End             string  `json:"end"`
	PriceAmount     float64 `json:"price_amount"`
	PriceCurrency   string  `json:"price_currency"`
}
type CreatePropertyRequest struct {
	OrganizationID string `json:"organization_id"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	Type           string `json:"type"`
}
type UpdatePropertyRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
	Type string `json:"type"`
}
type CreateUnitTypeRequest struct {
	PropertyID    string   `json:"property_id"`
	Name          string   `json:"name"`
	Code          string   `json:"code"`
	TotalQuantity int      `json:"total_quantity"`
	BasePrice     float64  `json:"base_price"`
	MaxOccupancy  int      `json:"max_occupancy"`
	MaxAdults     int      `json:"max_adults"`
	MaxChildren   int      `json:"max_children"`
	Amenities     []string `json:"amenities"`
}
type UpdateUnitTypeRequest struct {
	Name          string   `json:"name"`
	Code          string   `json:"code"`
	TotalQuantity int      `json:"total_quantity"`
	BasePrice     float64  `json:"base_price"`
	MaxOccupancy  int      `json:"max_occupancy"`
	MaxAdults     int      `json:"max_adults"`
	MaxChildren   int      `json:"max_children"`
	Amenities     []string `json:"amenities"`
}
type CreateUnitRequest struct {
	PropertyID string `json:"property_id"`
	UnitTypeID string `json:"unit_type_id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
}
type UpdateUnitRequest struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
