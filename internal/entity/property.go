package entity

type Property struct {
	BaseEntity
	
	OrganizationID string `json:"organization_id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	Type    string `json:"type"`
}

type CreatePropertyRequest struct {
	OrganizationID string `json:"organization_id"`
	Name string `json:"name"`
	Code string `json:"code"`
	Type string `json:"type"`
}

type UpdatePropertyRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
	Type string `json:"type"`
}
