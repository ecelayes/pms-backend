package domain

import "errors"

var ErrInvalidRole = errors.New("invalid role")

// UserRole is a value object representing the role of a user in the system.
type UserRole string

const (
	RoleUser       UserRole = "user"
	RoleSuperAdmin UserRole = "super_admin"
)

func NewUserRole(s string) (UserRole, error) {
	switch UserRole(s) {
	case RoleUser, RoleSuperAdmin:
		return UserRole(s), nil
	default:
		return "", ErrInvalidRole
	}
}

func (r UserRole) String() string       { return string(r) }
func (r UserRole) IsSuperAdmin() bool   { return r == RoleSuperAdmin }
func (r UserRole) IsUser() bool         { return r == RoleUser }
