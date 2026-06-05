package http

import (
	"context"

	"github.com/ecelayes/pms-backend/internal/iam/domain"
)

// AuthService is the contract for authentication operations.
type AuthService interface {
	Login(ctx context.Context, email, password string) (string, error)
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

// UserService is the contract for user management operations.
type UserService interface {
	Register(ctx context.Context, orgID, email, password, role, firstName, lastName, phone string) (string, error)
	GetAll(ctx context.Context, orgID string) ([]*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, id string, role, firstName, lastName, phone string) error
	Delete(ctx context.Context, id string) error
}

// OrganizationService is the contract for organization management operations.
type OrganizationService interface {
	Create(ctx context.Context, name, code string) (string, error)
	GetAll(ctx context.Context) ([]*domain.Organization, error)
	GetByID(ctx context.Context, id string) (*domain.Organization, error)
	Update(ctx context.Context, id, name, code string) error
	Delete(ctx context.Context, id string) error
}
