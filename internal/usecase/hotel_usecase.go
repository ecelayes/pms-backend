package usecase

import (
	"context"
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

func (uc *HotelUseCase) Create(ctx context.Context, req entity.CreateHotelRequest) (string, error) {
	if req.OrganizationID == "" {
		return "", entity.ErrInvalidInput
	}
	if req.Name == "" || req.Code == "" {
		return "", entity.ErrInvalidInput
	}
	if len(req.Code) < 3 || len(req.Code) > 5 {
		return "", entity.ErrInvalidInput
	}
	
	hotelID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid v7: %w", err)
	}

	hotel := entity.Hotel{
		BaseEntity: entity.BaseEntity{
			ID: hotelID.String(),
		},
		OrganizationID: req.OrganizationID, 
		Name:           req.Name,
		Code:           strings.ToUpper(req.Code),
	}

	return uc.repo.Create(ctx, hotel)
}

func (uc *HotelUseCase) ListByOrganization(ctx context.Context, orgID string, pagination entity.PaginationRequest) ([]entity.Hotel, int64, error) {
	if orgID == "" {
		return nil, 0, entity.ErrInvalidInput
	}
	return uc.repo.ListByOrganization(ctx, orgID, pagination)
}

func (uc *HotelUseCase) GetByID(ctx context.Context, id string) (*entity.Hotel, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, entity.ErrRecordNotFound
	}
	return uc.repo.GetByID(ctx, id)
}

func (uc *HotelUseCase) Update(ctx context.Context, id string, req entity.UpdateHotelRequest) error {
	if _, err := uuid.Parse(id); err != nil {
		return entity.ErrRecordNotFound
	}
	if req.Name == "" && req.Code == "" {
		return entity.ErrInvalidInput
	}
	if req.Code != "" {
		req.Code = strings.ToUpper(req.Code)
	}
	return uc.repo.Update(ctx, id, req)
}

func (uc *HotelUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return entity.ErrRecordNotFound
	}
	return uc.repo.Delete(ctx, id)
}
