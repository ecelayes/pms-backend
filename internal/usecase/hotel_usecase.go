package usecase

import (
	"context"
	"errors"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
)

type HotelUseCase struct {
	repo *repository.HotelRepository
}

func NewHotelUseCase(repo *repository.HotelRepository) *HotelUseCase {
	return &HotelUseCase{repo: repo}
}

func (uc *HotelUseCase) Create(ctx context.Context, ownerID string, req entity.CreateHotelRequest) (string, error) {
	if req.Name == "" || req.Code == "" {
		return "", errors.New("name and code are required")
	}
	if len(req.Code) < 3 || len(req.Code) > 5 {
		return "", errors.New("code must be between 3 and 5 characters")
	}
	return uc.repo.Create(ctx, ownerID, req)
}
