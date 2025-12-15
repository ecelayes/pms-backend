package usecase

import (
	"context"
	"errors"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
	"github.com/ecelayes/pms-backend/internal/service"
	"github.com/ecelayes/pms-backend/pkg/auth"
)

type AuthUseCase struct {
	db           *pgxpool.Pool
	userRepo     *repository.UserRepository
	emailService *service.EmailService
	logger       *zap.Logger
}

func NewAuthUseCase(
	db *pgxpool.Pool, 
	userRepo *repository.UserRepository, 
	emailService *service.EmailService,
	logger *zap.Logger,
) *AuthUseCase {
	return &AuthUseCase{
		db:           db,
		userRepo:     userRepo,
		emailService: emailService,
		logger:       logger,
	}
}

func (uc *AuthUseCase) Login(ctx context.Context, req entity.AuthRequest) (string, error) {
	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", err
	}

	if !auth.CheckPassword(req.Password, user.Password) {
		return "", entity.ErrInvalidCredentials
	}

	return auth.GenerateToken(user.ID, user.Role, user.Salt)
}

func (uc *AuthUseCase) GetUserSalt(ctx context.Context, userID string) (string, error) {
	return uc.userRepo.GetSaltByID(ctx, userID)
}

func (uc *AuthUseCase) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidCredentials) {
			uc.logger.Warn("password reset requested for non-existent email", 
				zap.String("email", email),
			)
			return nil 
		}
		uc.logger.Error("database error looking up user for reset", 
			zap.Error(err),
		)
		return err
	}

	token, err := auth.GenerateResetToken(user.ID, user.Salt)
	if err != nil {
		uc.logger.Error("failed to generate reset token", zap.Error(err))
		return err
	}

	userName := user.FirstName
    if userName == "" {
        userName = "Usuario"
    } 

	if err := uc.emailService.SendPasswordReset(user.Email, userName, token); err != nil {
		uc.logger.Error("failed to send password reset email",
			zap.String("user_id", user.ID),
			zap.String("email", user.Email),
			zap.Error(err),
		)
		return nil 
	}
	
	uc.logger.Info("password reset email sent successfully", 
		zap.String("user_id", user.ID),
	)

	return nil
}

func (uc *AuthUseCase) ResetPassword(ctx context.Context, req entity.ResetPasswordRequest) error {
	claims, err := auth.ParseTokenClaimsUnsafe(req.Token)
	if err != nil {
		return errors.New("invalid token format")
	}

	if claims.Purpose != auth.PurposeReset {
		return errors.New("invalid token type")
	}

	currentSalt, err := uc.userRepo.GetSaltByID(ctx, claims.UserID)
	if err != nil {
		return errors.New("invalid token or user not found")
	}

	if _, err := auth.ValidateSignature(req.Token, currentSalt); err != nil {
		return errors.New("token expired or invalid")
	}

	newPasswordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil { return err }
	
	newSalt, err := auth.GenerateRandomSalt()
	if err != nil { return err }

	return uc.userRepo.UpdatePassword(ctx, claims.UserID, newPasswordHash, newSalt)
}
