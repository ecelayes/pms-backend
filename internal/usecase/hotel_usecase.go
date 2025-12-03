package usecase

import (
	"context"
	"errors"
	"strings"
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
	req.Code = strings.ToUpper(req.Code)
	return uc.repo.Create(ctx, ownerID, req)
}

func (uc *HotelUseCase) Update(ctx context.Context, id string, req entity.UpdateHotelRequest) error {
	if req.Name == "" && req.Code == "" {
		return errors.New("nothing to update")
	}
	if req.Code != "" {
		req.Code = strings.ToUpper(req.Code)
	}
	return uc.repo.Update(ctx, id, req)
}

func (uc *HotelUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *HotelUseCase) ListMine(ctx context.Context, ownerID string) ([]entity.Hotel, error) {
	return uc.repo.ListByOwner(ctx, ownerID)
}
