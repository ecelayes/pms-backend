package http

import (
	"context"

	"github.com/ecelayes/pms-backend/internal/iam/domain"
)

// mockAuthService is a configurable mock for testing the HTTP handler in isolation.
type mockAuthService struct {
	loginToken       string
	loginErr         error
	resetErr         error
	resetPasswordErr error
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (string, error) {
	return m.loginToken, m.loginErr
}
func (m *mockAuthService) RequestPasswordReset(ctx context.Context, email string) error {
	return m.resetErr
}
func (m *mockAuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	return m.resetPasswordErr
}

type mockUserService struct {
	registerResult string
	registerErr    error
	getAllResult   []*domain.User
	getAllErr      error
	getByIDResult  *domain.User
	getByIDErr     error
	updateErr      error
	deleteErr      error
}

func (m *mockUserService) Register(ctx context.Context, orgID, email, password, role, firstName, lastName, phone string) (string, error) {
	return m.registerResult, m.registerErr
}
func (m *mockUserService) GetAll(ctx context.Context, orgID string) ([]*domain.User, error) {
	return m.getAllResult, m.getAllErr
}
func (m *mockUserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return m.getByIDResult, m.getByIDErr
}
func (m *mockUserService) Update(ctx context.Context, id string, role, firstName, lastName, phone string) error {
	return m.updateErr
}
func (m *mockUserService) Delete(ctx context.Context, id string) error {
	return m.deleteErr
}

type mockOrganizationService struct {
	createResult  string
	createErr     error
	getAllResult  []*domain.Organization
	getAllErr     error
	getByIDResult *domain.Organization
	getByIDErr    error
	updateErr     error
	deleteErr     error
}

func (m *mockOrganizationService) Create(ctx context.Context, name, code string) (string, error) {
	return m.createResult, m.createErr
}
func (m *mockOrganizationService) GetAll(ctx context.Context) ([]*domain.Organization, error) {
	return m.getAllResult, m.getAllErr
}
func (m *mockOrganizationService) GetByID(ctx context.Context, id string) (*domain.Organization, error) {
	return m.getByIDResult, m.getByIDErr
}
func (m *mockOrganizationService) Update(ctx context.Context, id, name, code string) error {
	return m.updateErr
}
func (m *mockOrganizationService) Delete(ctx context.Context, id string) error {
	return m.deleteErr
}
