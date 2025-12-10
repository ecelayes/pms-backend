package entity

const (
	RoleUser       = "user"
	RoleSuperAdmin = "super_admin"
)

type User struct {
	BaseEntity
	Email    string `json:"email"`
	Password string `json:"-"`
	Salt     string `json:"-"`
	Role     string `json:"role"`
}

type CreateUserRequest struct {
	OrganizationID string `json:"organization_id"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	Name      		 string `json:"name"`
	Role           string `json:"role"`
}

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterOwnerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	OrgName  string `json:"org_name"`
}

type AuthResponse struct {
	Token string `json:"token"`
}
