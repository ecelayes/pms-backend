package entity

const (
	RoleUser       = "user"
	RoleSuperAdmin = "super_admin"
)

type User struct {
	BaseEntity
	Email     string `json:"email"`
	Password  string `json:"-"`
	Salt      string `json:"-"`
	Role      string `json:"role"`
	
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

type CreateUserRequest struct {
	OrganizationID string `json:"organization_id"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	Role           string `json:"role"`
	
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Phone          string `json:"phone"`
}

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type RegisterOwnerRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	OrgName   string `json:"org_name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

type UpdateUserRequest struct {
	Email     string `json:"email"`
	Role      string `json:"role"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

type AuthResponse struct {
	Token string `json:"token"`
}
