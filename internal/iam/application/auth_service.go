package application

import (
	"context"
	"errors"
	"github.com/ecelayes/pms-backend/internal/iam/domain"
	"github.com/ecelayes/pms-backend/pkg/auth"
)

var (
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidTokenPurpose   = errors.New("invalid token purpose")
	ErrInvalidOrExpiredToken = errors.New("invalid or expired token")
	ErrWeakPassword          = errors.New("password does not meet strength requirements")
)

type EmailService interface {
	SendPasswordReset(toEmail, userName, token string) error
}

// AuthService orchestrates authentication and password reset flows.
//
// SECURITY: All cryptographic operations are delegated to interfaces
// (PasswordHasher, TokenGenerator, RandomSaltGenerator) defined in pkg/auth.
// The production wiring (in module.go) uses BcryptPasswordHasher,
// JWTTokenGenerator, and CryptoRandomSaltGenerator - all using
// industry-standard, slow, salted algorithms. The interfaces enable
// testability without weakening production security guarantees.
type AuthService struct {
	userRepo        domain.UserRepository
	orgRepo         domain.OrganizationRepository
	emailService    EmailService
	passwordHasher  auth.PasswordHasher
	tokenGenerator  auth.TokenGenerator
	saltGenerator   auth.RandomSaltGenerator
}

func NewAuthService(
	userRepo domain.UserRepository,
	orgRepo domain.OrganizationRepository,
	emailService EmailService,
	passwordHasher auth.PasswordHasher,
	tokenGenerator auth.TokenGenerator,
	saltGenerator auth.RandomSaltGenerator,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		orgRepo:        orgRepo,
		emailService:   emailService,
		passwordHasher: passwordHasher,
		tokenGenerator: tokenGenerator,
		saltGenerator:  saltGenerator,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}
	// SECURITY: Constant-time comparison via bcrypt
	if !s.passwordHasher.Verify(password, user.Password()) {
		return "", ErrInvalidCredentials
	}
	org, err := s.orgRepo.FindByUserID(ctx, user.ID())
	if err != nil {
		return "", err
	}
	orgID := ""
	if org != nil {
		orgID = org.ID()
	}
	role := string(user.Role())
	return s.tokenGenerator.GenerateAuthToken(user.ID(), orgID, role, user.Salt())
}

func (s *AuthService) GetUserSalt(ctx context.Context, userID string) (string, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", ErrUserNotFound
	}
	return u.Salt(), nil
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// SECURITY: silently return nil to avoid leaking user existence via timing/error
		return nil //nolint:nilerr
	}
	if user == nil {
		return nil
	}
	token, err := s.tokenGenerator.GenerateResetToken(user.ID(), user.Salt())
	if err != nil {
		return err
	}
	go func() {
		_ = s.emailService.SendPasswordReset(user.Email(), user.FirstName(), token)
	}()
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	claims, err := s.tokenGenerator.ParseUnsafe(token)
	if err != nil {
		return err
	}
	if claims.Purpose != auth.PurposeReset {
		return ErrInvalidTokenPurpose
	}
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	// SECURITY: Verify signature BEFORE trusting the claims
	if _, err := s.tokenGenerator.VerifySignature(token, user.Salt()); err != nil {
		return ErrInvalidOrExpiredToken
	}
	if _, err := domain.NewPassword(newPassword); err != nil {
		return ErrWeakPassword
	}
	hashed, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return err
	}
	newSalt, err := s.saltGenerator.Generate()
	if err != nil {
		return err
	}
	user.ChangePassword(hashed, newSalt)
	return s.userRepo.Save(ctx, user)
}
