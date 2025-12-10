package entity

type Organization struct {
	BaseEntity
	Name string `json:"name"`
	Code string `json:"code"`
}

type OrganizationMember struct {
	BaseEntity
	OrganizationID string `json:"organization_id"`
	UserID         string `json:"user_id"`
	Role           string `json:"role"`
}

type CreateOrganizationRequest struct {
	Name string `json:"name"`
}

type UpdateOrganizationRequest struct {
	Name string `json:"name"`
}

const (
	OrgRoleOwner   = "owner"
	OrgRoleManager = "manager"
	OrgRoleStaff   = "staff"
)
