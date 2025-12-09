package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
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
	
	hotelID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid v7: %w", err)
	}

	hotel := entity.Hotel{
		BaseEntity: entity.BaseEntity{
			ID: hotelID.String(),
		},
		OwnerID: ownerID,
		Name:    req.Name,
		Code:    strings.ToUpper(req.Code),
	}

	return uc.repo.Create(ctx, hotel)
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
