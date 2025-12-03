package usecase

import (
	"context"

	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
	"github.com/ecelayes/pms-backend/pkg/auth"
)

type AuthUseCase struct {
	repo *repository.UserRepository
}

func NewAuthUseCase(repo *repository.UserRepository) *AuthUseCase {
	return &AuthUseCase{repo: repo}
}

func (uc *AuthUseCase) Register(ctx context.Context, req entity.AuthRequest) error {
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return err
	}
	
	userSalt, err := auth.GenerateRandomSalt()
	if err != nil {
		return err
	}

	return uc.repo.Create(ctx, req.Email, passwordHash, userSalt, "admin")
}

func (uc *AuthUseCase) Login(ctx context.Context, req entity.AuthRequest) (string, error) {
	user, err := uc.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", err
	}

	if !auth.CheckPassword(req.Password, user.Password) {
		return "", entity.ErrInvalidCredentials
	}

	return auth.GenerateToken(user.ID, user.Role, user.Salt)
}

func (uc *AuthUseCase) GetUserSalt(ctx context.Context, userID string) (string, error) {
	return uc.repo.GetSaltByID(ctx, userID)
}
