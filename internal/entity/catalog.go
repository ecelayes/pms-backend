package entity

type Amenity struct {
	BaseEntity
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type HotelService struct {
	BaseEntity
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type CreateCatalogRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type UpdateCatalogRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}
