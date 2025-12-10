package entity

type Hotel struct {
	BaseEntity
	
	OrganizationID string `json:"organization_id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
}

type CreateHotelRequest struct {
	OrganizationID string `json:"organization_id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type UpdateHotelRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}
