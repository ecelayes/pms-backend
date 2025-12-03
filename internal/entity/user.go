package entity

type User struct {
	BaseEntity
	
	Email    string `json:"email"`
	Password string `json:"-"`
	Salt     string `json:"-"`
	Role     string `json:"role"`
}

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}
