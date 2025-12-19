package entity

type Unit struct {
	BaseEntity
	
	PropertyID  string `json:"property_id"`
	UnitTypeID  string `json:"unit_type_id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
}

type CreateUnitRequest struct {
	PropertyID  string `json:"property_id"`
	UnitTypeID  string `json:"unit_type_id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
}

type UpdateUnitRequest struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
