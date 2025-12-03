package entity

type Hotel struct {
	BaseEntity
	
	OwnerID string `json:"owner_id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
}

type CreateHotelRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type UpdateHotelRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}
