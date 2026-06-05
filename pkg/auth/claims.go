package auth

import "github.com/golang-jwt/jwt/v5"

const (
	PurposeAuth  = "auth"
	PurposeReset = "reset"
)

// Claims represents the JWT claims used for authentication and password reset.
//
// SECURITY: The Purpose field MUST be set to PurposeAuth or PurposeReset.
// The PurposeAuth is required for normal API access. PurposeReset is only
// valid for the password reset flow. Tokens with any other purpose are rejected.
type Claims struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	Role           string `json:"role"`
	Purpose        string `json:"purpose"`
	jwt.RegisteredClaims
}
