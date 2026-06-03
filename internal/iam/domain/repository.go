package domain

import "context"

type UserRepository interface {
	Save(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindAllByOrganization(ctx context.Context, orgID string) ([]*User, error)
	DeleteUser(ctx context.Context, id string) error
	EnsureGuest(ctx context.Context, u *User) error
}
type OrganizationRepository interface {
	Save(ctx context.Context, o *Organization) error
	FindByID(ctx context.Context, id string) (*Organization, error)
	AddMember(ctx context.Context, orgID, userID, role string) error
	FindMemberRole(ctx context.Context, orgID, userID string) (string, error)
	FindByUserID(ctx context.Context, userID string) (*Organization, error)
	FindAll(ctx context.Context) ([]*Organization, error)
	DeleteOrganization(ctx context.Context, id string) error
}
