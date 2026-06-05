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
	svc := newAuthServiceForTest(nil, nil, nil)
	if svc == nil {
		t.Error("Expected non-nil service")
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{findByEmailResult: nil}
	svc := newAuthServiceForTest(repo, nil, nil)

	_, err := svc.Login(context.Background(), "test@test.com", "password")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	user := domain.ReconstituteUser("user-1", "test@test.com", "hash", "salt", string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	repo := &mockUserRepo{findByEmailResult: user}
	hasher := &mockPasswordHasher{hashResult: "hashed", verifyResult: false}
	_, tg, sg := newStandardMocks()
	svc := NewAuthService(repo, nil, nil, hasher, tg, sg)

	_, err := svc.Login(context.Background(), "test@test.com", "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	// Hash the password correctly
	hashed, _ := auth.NewBcryptPasswordHasher().Hash("correctpassword")
	salt, _ := auth.NewCryptoRandomSaltGenerator().Generate()
	user := domain.ReconstituteUser("user-1", "test@test.com", hashed, salt, string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	
	repo := &mockUserRepo{findByEmailResult: user}
	orgRepo := &mockOrgRepo{findByUserIDResult: domain.ReconstituteOrganization("org-1", "Test Org", "test.com", time.Now())}
	svc := newAuthServiceForTest(repo, orgRepo, nil)

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
	svc := newAuthServiceForTest(repo, nil, nil)

	_, err := svc.Login(context.Background(), "test@test.com", "password")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestAuthService_GetUserSalt_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{findByIDResult: nil}
	svc := newAuthServiceForTest(repo, nil, nil)

	_, err := svc.GetUserSalt(context.Background(), "non-existent")
	if err == nil {
		t.Error("Expected error for not found user")
	}
}

func TestAuthService_GetUserSalt_Success(t *testing.T) {
	salt := "test-salt-123"
	user := domain.ReconstituteUser("user-1", "test@test.com", "hash", salt, string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	repo := &mockUserRepo{findByIDResult: user}
	svc := newAuthServiceForTest(repo, nil, nil)

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
	svc := newAuthServiceForTest(repo, nil, nil)

	_, err := svc.GetUserSalt(context.Background(), "user-1")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestAuthService_RequestPasswordReset_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{findByEmailResult: nil}
	svc := newAuthServiceForTest(repo, nil, nil)

	// Should return nil even if user not found (security best practice)
	err := svc.RequestPasswordReset(context.Background(), "non-existent@test.com")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestAuthService_RequestPasswordReset_UserFound(t *testing.T) {
	hashed, _ := auth.NewBcryptPasswordHasher().Hash("password")
	salt, _ := auth.NewCryptoRandomSaltGenerator().Generate()
	user := domain.ReconstituteUser("user-1", "test@test.com", hashed, salt, string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	repo := &mockUserRepo{findByEmailResult: user}
	emailSvc := &mockEmailService{}
	svc := newAuthServiceForTest(repo, nil, emailSvc)

	// This will spawn a goroutine for email - we can't easily test that
	err := svc.RequestPasswordReset(context.Background(), "test@test.com")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestAuthService_ResetPassword_InvalidToken(t *testing.T) {
	svc := newAuthServiceForTest(nil, nil, nil)

	err := svc.ResetPassword(context.Background(), "invalid-token", "newpassword")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestAuthService_ResetPassword_TokenPurposeMismatch(t *testing.T) {
	// Create a token with wrong purpose
	hashed, _ := auth.NewBcryptPasswordHasher().Hash("password")
	salt, _ := auth.NewCryptoRandomSaltGenerator().Generate()
	user := domain.ReconstituteUser("user-1", "test@test.com", hashed, salt, string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	repo := &mockUserRepo{findByIDResult: user}
	svc := newAuthServiceForTest(repo, nil, nil)

	// Generate a login token instead of reset token (PurposeAuth instead of PurposeReset)
	token, _ := auth.NewJWTTokenGenerator().GenerateAuthToken(user.ID(), "", string(auth.PurposeAuth), user.Salt())

	err := svc.ResetPassword(context.Background(), token, "newpassword")
	if err == nil {
		t.Error("Expected error for wrong token purpose")
	}
}

func TestAuthService_ResetPassword_UserNotFound(t *testing.T) {
	svc := newAuthServiceForTest(nil, nil, nil)

	// This will fail on ParseTokenClaimsUnsafe since token is invalid format
	err := svc.ResetPassword(context.Background(), "some-invalid-token-format", "newpassword")
	// Should error on invalid token format
	if err == nil {
		t.Error("Expected error for invalid token format")
	}
}

func TestAuthService_ResetPassword_Success(t *testing.T) {
	hashed := "hashed"
	salt := "salt"
	user := domain.ReconstituteUser("user-1", "test@test.com", hashed, salt, string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	repo := &mockUserRepo{findByIDResult: user}
	hasher := &mockPasswordHasher{hashResult: "newhash", verifyResult: true}
	tg := &mockTokenGenerator{
		generateResetResult: "reset-token",
		parseUnsafeResult:   &auth.Claims{UserID: user.ID(), Purpose: auth.PurposeReset},
		verifySigResult:     &auth.Claims{UserID: user.ID(), Purpose: auth.PurposeReset},
	}
	sg := &mockSaltGenerator{generateRes: "newsalt"}
	svc := NewAuthService(repo, nil, nil, hasher, tg, sg)

	// Pass a valid-format token
	token, _ := auth.NewJWTTokenGenerator().GenerateResetToken(user.ID(), user.Salt())
	_ = token
	err := svc.ResetPassword(context.Background(), "valid-token", "newpassword")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestAuthService_ResetPassword_InvalidSignature(t *testing.T) {
	hashed := "hashed"
	salt := "salt"
	user := domain.ReconstituteUser("user-1", "test@test.com", hashed, salt, string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	repo := &mockUserRepo{findByIDResult: user}
	hasher := &mockPasswordHasher{hashResult: "hashed", verifyResult: true}
	tg := &mockTokenGenerator{
		parseUnsafeResult: &auth.Claims{UserID: user.ID(), Purpose: auth.PurposeReset},
		verifySigErr:      errors.New("invalid signature"),
	}
	sg := &mockSaltGenerator{generateRes: "salt"}
	svc := NewAuthService(repo, nil, nil, hasher, tg, sg)

	err := svc.ResetPassword(context.Background(), "valid-token", "newpassword")
	if err == nil {
		t.Error("Expected error for invalid signature")
	}
}

func TestAuthService_ResetPassword_HashError(t *testing.T) {
	hashed := "hashed"
	salt := "salt"
	user := domain.ReconstituteUser("user-1", "test@test.com", hashed, salt, string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	repo := &mockUserRepo{findByIDResult: user}
	hasher := &mockPasswordHasher{hashErr: errors.New("hash error"), verifyResult: true}
	tg := &mockTokenGenerator{
		parseUnsafeResult: &auth.Claims{UserID: user.ID(), Purpose: auth.PurposeReset},
		verifySigResult:   &auth.Claims{UserID: user.ID(), Purpose: auth.PurposeReset},
	}
	sg := &mockSaltGenerator{generateRes: "salt"}
	svc := NewAuthService(repo, nil, nil, hasher, tg, sg)

	err := svc.ResetPassword(context.Background(), "valid-token", "newpassword")
	if err == nil {
		t.Error("Expected hash error")
	}
}

func TestAuthService_ResetPassword_SaltError(t *testing.T) {
	hashed := "hashed"
	salt := "salt"
	user := domain.ReconstituteUser("user-1", "test@test.com", hashed, salt, string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	repo := &mockUserRepo{findByIDResult: user}
	hasher := &mockPasswordHasher{hashResult: "hashed", verifyResult: true}
	tg := &mockTokenGenerator{
		parseUnsafeResult: &auth.Claims{UserID: user.ID(), Purpose: auth.PurposeReset},
		verifySigResult:   &auth.Claims{UserID: user.ID(), Purpose: auth.PurposeReset},
	}
	sg := &mockSaltGenerator{generateErr: errors.New("salt error")}
	svc := NewAuthService(repo, nil, nil, hasher, tg, sg)

	err := svc.ResetPassword(context.Background(), "valid-token", "newpassword")
	if err == nil {
		t.Error("Expected salt error")
	}
}

func TestAuthService_ResetPassword_SaveError(t *testing.T) {
	hashed := "hashed"
	salt := "salt"
	user := domain.ReconstituteUser("user-1", "test@test.com", hashed, salt, string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	repo := &mockUserRepo{findByIDResult: user, saveErr: errors.New("save error")}
	hasher := &mockPasswordHasher{hashResult: "hashed", verifyResult: true}
	tg := &mockTokenGenerator{
		parseUnsafeResult: &auth.Claims{UserID: user.ID(), Purpose: auth.PurposeReset},
		verifySigResult:   &auth.Claims{UserID: user.ID(), Purpose: auth.PurposeReset},
	}
	sg := &mockSaltGenerator{generateRes: "salt"}
	svc := NewAuthService(repo, nil, nil, hasher, tg, sg)

	err := svc.ResetPassword(context.Background(), "valid-token", "newpassword")
	if err == nil {
		t.Error("Expected save error")
	}
}

func TestAuthService_ResetPassword_FindByIDError(t *testing.T) {
	repo := &mockUserRepo{findByIDErr: errors.New("find error")}
	hasher := &mockPasswordHasher{hashResult: "hashed", verifyResult: true}
	tg := &mockTokenGenerator{
		parseUnsafeResult: &auth.Claims{UserID: "user-1", Purpose: auth.PurposeReset},
	}
	sg := &mockSaltGenerator{generateRes: "salt"}
	svc := NewAuthService(repo, nil, nil, hasher, tg, sg)

	err := svc.ResetPassword(context.Background(), "valid-token", "newpassword")
	if err == nil {
		t.Error("Expected find error")
	}
}

func TestAuthService_ResetPassword_UserNil(t *testing.T) {
	repo := &mockUserRepo{findByIDResult: nil}
	hasher := &mockPasswordHasher{hashResult: "hashed", verifyResult: true}
	tg := &mockTokenGenerator{
		parseUnsafeResult: &auth.Claims{UserID: "user-1", Purpose: auth.PurposeReset},
	}
	sg := &mockSaltGenerator{generateRes: "salt"}
	svc := NewAuthService(repo, nil, nil, hasher, tg, sg)

	err := svc.ResetPassword(context.Background(), "valid-token", "newpassword")
	if err == nil {
		t.Error("Expected user not found error")
	}
}

func TestAuthService_RequestPasswordReset_TokenError(t *testing.T) {
	user := domain.ReconstituteUser("user-1", "test@test.com", "hash", "salt", string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	repo := &mockUserRepo{findByEmailResult: user}
	hasher := &mockPasswordHasher{hashResult: "hashed", verifyResult: true}
	tg := &mockTokenGenerator{generateResetErr: errors.New("token error")}
	sg := &mockSaltGenerator{generateRes: "salt"}
	svc := NewAuthService(repo, nil, nil, hasher, tg, sg)

	err := svc.RequestPasswordReset(context.Background(), "test@test.com")
	if err == nil {
		t.Error("Expected token error")
	}
}

func TestAuthService_Login_OrgRepoError(t *testing.T) {
	user := domain.ReconstituteUser("user-1", "test@test.com", "hash", "salt", string(domain.RoleSuperAdmin), "John", "Doe", "", time.Now())
	repo := &mockUserRepo{findByEmailResult: user}
	orgRepo := &mockOrgRepo{findByUserIDErr: errors.New("org repo error")}
	hasher := &mockPasswordHasher{hashResult: "hashed", verifyResult: true}
	_, tg, sg := newStandardMocks()
	svc := NewAuthService(repo, orgRepo, nil, hasher, tg, sg)

	_, err := svc.Login(context.Background(), "test@test.com", "password")
	if err == nil {
		t.Error("Expected org repo error")
	}
}

func TestAuthService_RequestPasswordReset_FindError(t *testing.T) {
	repo := &mockUserRepo{findByEmailErr: errors.New("find error")}
	hasher := &mockPasswordHasher{hashResult: "hashed", verifyResult: true}
	_, tg, sg := newStandardMocks()
	svc := NewAuthService(repo, nil, nil, hasher, tg, sg)

	err := svc.RequestPasswordReset(context.Background(), "test@test.com")
	if err != nil {
		t.Errorf("Expected nil error (graceful), got: %v", err)
	}
}

func TestAuthService_ResetPassword_ParseUnsafeError(t *testing.T) {
	hasher := &mockPasswordHasher{hashResult: "hashed", verifyResult: true}
	tg := &mockTokenGenerator{parseUnsafeErr: errors.New("parse error")}
	sg := &mockSaltGenerator{generateRes: "salt"}
	svc := NewAuthService(nil, nil, nil, hasher, tg, sg)

	err := svc.ResetPassword(context.Background(), "malformed-token", "newpassword")
	if err == nil {
		t.Error("Expected parse error")
	}
}
