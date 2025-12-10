package usecase

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
	"github.com/ecelayes/pms-backend/pkg/auth"
)

type AuthUseCase struct {
	db       *pgxpool.Pool
	userRepo *repository.UserRepository
}

func NewAuthUseCase(db *pgxpool.Pool, userRepo *repository.UserRepository) *AuthUseCase {
	return &AuthUseCase{
		db:       db,
		userRepo: userRepo,
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
