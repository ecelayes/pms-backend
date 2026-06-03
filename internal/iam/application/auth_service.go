package application

import (
	"context"
	"errors"
	"github.com/ecelayes/pms-backend/internal/iam/domain"
	"github.com/ecelayes/pms-backend/pkg/auth"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type EmailService interface {
	SendPasswordReset(toEmail, userName, token string) error
}
type AuthService struct {
	userRepo     domain.UserRepository
	orgRepo      domain.OrganizationRepository
	emailService EmailService
}

func NewAuthService(userRepo domain.UserRepository, orgRepo domain.OrganizationRepository, emailService EmailService) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		orgRepo:      orgRepo,
		emailService: emailService,
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
	if !auth.CheckPassword(password, user.Password()) {
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
	return auth.GenerateToken(user.ID(), orgID, role, user.Salt())
}
func (s *AuthService) GetUserSalt(ctx context.Context, userID string) (string, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", errors.New("user not found")
	}
	return u.Salt(), nil
}
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil
	}
	if user == nil {
		return nil
	}
	token, err := auth.GenerateToken(user.ID(), "", string(auth.PurposeReset), user.Salt())
	if err != nil {
		return err
	}
	go func() {
		_ = s.emailService.SendPasswordReset(user.Email(), user.FirstName(), token)
	}()
	return nil
}
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	claims, err := auth.ParseTokenClaimsUnsafe(token)
	if err != nil {
		return err
	}
	if claims.Purpose != auth.PurposeReset {
		return errors.New("invalid token purpose")
	}
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	if _, err := auth.ValidateSignature(token, user.Salt()); err != nil {
		return errors.New("invalid or expired token")
	}
	hashed, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	newSalt, err := auth.GenerateRandomSalt()
	if err != nil {
		return err
	}
	user.ChangePassword(hashed, newSalt)
	return s.userRepo.Save(ctx, user)
}
