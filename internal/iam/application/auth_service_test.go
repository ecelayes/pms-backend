package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ecelayes/pms-backend/internal/iam/domain"
	"github.com/ecelayes/pms-backend/pkg/auth"
)


type mockUserRepo struct {
	findByEmailResult *domain.User
	findByEmailErr    error
	findByIDResult    *domain.User
	findByIDErr       error
	saveErr           error
}

func (m *mockUserRepo) Save(ctx context.Context, u *domain.User) error {
	return m.saveErr
}
func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	return m.findByIDResult, nil
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.findByEmailErr != nil {
		return nil, m.findByEmailErr
	}
	return m.findByEmailResult, nil
}
func (m *mockUserRepo) FindAllByOrganization(ctx context.Context, orgID string) ([]*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepo) DeleteUser(ctx context.Context, id string) error {
	return nil
}
func (m *mockUserRepo) EnsureGuest(ctx context.Context, u *domain.User) error {
	return nil
}

type mockOrgRepo struct {
	findByUserIDResult *domain.Organization
	findByUserIDErr     error
}

func (m *mockOrgRepo) Save(ctx context.Context, o *domain.Organization) error {
	return nil
}
func (m *mockOrgRepo) FindByID(ctx context.Context, id string) (*domain.Organization, error) {
	return nil, nil
}
func (m *mockOrgRepo) FindByUserID(ctx context.Context, userID string) (*domain.Organization, error) {
	if m.findByUserIDErr != nil {
		return nil, m.findByUserIDErr
	}
	return m.findByUserIDResult, nil
}
func (m *mockOrgRepo) FindByDomain(ctx context.Context, domain string) (*domain.Organization, error) {
	return nil, nil
}
func (m *mockOrgRepo) AddMember(ctx context.Context, orgID, userID, role string) error {
	return nil
}
func (m *mockOrgRepo) FindMemberRole(ctx context.Context, orgID, userID string) (string, error) {
	return "", nil
}
func (m *mockOrgRepo) FindAll(ctx context.Context) ([]*domain.Organization, error) {
	return nil, nil
}
func (m *mockOrgRepo) DeleteOrganization(ctx context.Context, id string) error {
	return nil
}

type mockEmailService struct {
	sendErr error
}

func (m *mockEmailService) SendPasswordReset(toEmail, userName, token string) error {
	return m.sendErr
}


func TestNewAuthService(t *testing.T) {
	svc := NewAuthService(nil, nil, nil)
	if svc == nil {
		t.Error("Expected non-nil service")
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{findByEmailResult: nil}
	svc := NewAuthService(repo, nil, nil)

	_, err := svc.Login(context.Background(), "test@test.com", "password")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	user := domain.ReconstituteUser("user-1", "test@test.com", "hash", "salt", string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	repo := &mockUserRepo{findByEmailResult: user}
	svc := NewAuthService(repo, nil, nil)

	_, err := svc.Login(context.Background(), "test@test.com", "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	// Hash the password correctly
	hashed, _ := auth.HashPassword("correctpassword")
	salt, _ := auth.GenerateRandomSalt()
	user := domain.ReconstituteUser("user-1", "test@test.com", hashed, salt, string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	
	repo := &mockUserRepo{findByEmailResult: user}
	orgRepo := &mockOrgRepo{findByUserIDResult: domain.ReconstituteOrganization("org-1", "Test Org", "test.com", time.Now())}
	svc := NewAuthService(repo, orgRepo, nil)

	token, err := svc.Login(context.Background(), "test@test.com", "correctpassword")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if token == "" {
		t.Error("Expected non-empty token")
	}
}

func TestAuthService_Login_UserRepoError(t *testing.T) {
	repo := &mockUserRepo{findByEmailErr: errors.New("repo error")}
	svc := NewAuthService(repo, nil, nil)

	_, err := svc.Login(context.Background(), "test@test.com", "password")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestAuthService_GetUserSalt_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{findByIDResult: nil}
	svc := NewAuthService(repo, nil, nil)

	_, err := svc.GetUserSalt(context.Background(), "non-existent")
	if err == nil {
		t.Error("Expected error for not found user")
	}
}

func TestAuthService_GetUserSalt_Success(t *testing.T) {
	salt := "test-salt-123"
	user := domain.ReconstituteUser("user-1", "test@test.com", "hash", salt, string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	repo := &mockUserRepo{findByIDResult: user}
	svc := NewAuthService(repo, nil, nil)

	result, err := svc.GetUserSalt(context.Background(), "user-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != salt {
		t.Errorf("Expected salt '%s', got '%s'", salt, result)
	}
}

func TestAuthService_GetUserSalt_RepoError(t *testing.T) {
	repo := &mockUserRepo{findByIDErr: errors.New("repo error")}
	svc := NewAuthService(repo, nil, nil)

	_, err := svc.GetUserSalt(context.Background(), "user-1")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestAuthService_RequestPasswordReset_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{findByEmailResult: nil}
	svc := NewAuthService(repo, nil, nil)

	// Should return nil even if user not found (security best practice)
	err := svc.RequestPasswordReset(context.Background(), "non-existent@test.com")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestAuthService_RequestPasswordReset_UserFound(t *testing.T) {
	hashed, _ := auth.HashPassword("password")
	salt, _ := auth.GenerateRandomSalt()
	user := domain.ReconstituteUser("user-1", "test@test.com", hashed, salt, string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	repo := &mockUserRepo{findByEmailResult: user}
	emailSvc := &mockEmailService{}
	svc := NewAuthService(repo, nil, emailSvc)

	// This will spawn a goroutine for email - we can't easily test that
	err := svc.RequestPasswordReset(context.Background(), "test@test.com")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestAuthService_ResetPassword_InvalidToken(t *testing.T) {
	svc := NewAuthService(nil, nil, nil)

	err := svc.ResetPassword(context.Background(), "invalid-token", "newpassword")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestAuthService_ResetPassword_TokenPurposeMismatch(t *testing.T) {
	// Create a token with wrong purpose
	hashed, _ := auth.HashPassword("password")
	salt, _ := auth.GenerateRandomSalt()
	user := domain.ReconstituteUser("user-1", "test@test.com", hashed, salt, string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	repo := &mockUserRepo{findByIDResult: user}
	svc := NewAuthService(repo, nil, nil)

	// Generate a login token instead of reset token (PurposeAuth instead of PurposeReset)
	token, _ := auth.GenerateToken(user.ID(), "", string(auth.PurposeAuth), user.Salt())

	err := svc.ResetPassword(context.Background(), token, "newpassword")
	if err == nil {
		t.Error("Expected error for wrong token purpose")
	}
}

func TestAuthService_ResetPassword_UserNotFound(t *testing.T) {
	svc := NewAuthService(nil, nil, nil)

	// This will fail on ParseTokenClaimsUnsafe since token is invalid format
	err := svc.ResetPassword(context.Background(), "some-invalid-token-format", "newpassword")
	// Should error on invalid token format
	if err == nil {
		t.Error("Expected error for invalid token format")
	}
}
