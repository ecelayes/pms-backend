package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
)

type RoomUseCase struct {
	repo *repository.RoomRepository
}

func NewRoomUseCase(repo *repository.RoomRepository) *RoomUseCase {
	return &RoomUseCase{repo: repo}
}

func (uc *RoomUseCase) Create(ctx context.Context, req entity.CreateRoomTypeRequest) (string, error) {
	if req.HotelID == "" || req.Name == "" {
		return "", entity.ErrInvalidInput
	}
	if len(req.Code) != 3 {
		return "", entity.ErrInvalidInput
	}
	if req.TotalQuantity < 0 {
		return "", entity.ErrInvalidInput
	}
	if req.BasePrice < 0 {
		return "", entity.ErrInvalidInput
	}
	if req.MaxOccupancy <= 0 {
		return "", entity.ErrInvalidInput
	}
	if req.MaxAdults <= 0 {
		return "", entity.ErrInvalidInput
	}

	newID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid v7: %w", err)
	}

	rt := entity.RoomType{
		BaseEntity: entity.BaseEntity{
			ID: newID.String(),
		},
		HotelID:       req.HotelID,
		Name:          req.Name,
		Code:          strings.ToUpper(req.Code),
		TotalQuantity: req.TotalQuantity,
		BasePrice:     req.BasePrice,
		MaxOccupancy:  req.MaxOccupancy,
		MaxAdults:     req.MaxAdults,
		MaxChildren:   req.MaxChildren,
		Amenities:     req.Amenities,
	}

	return uc.repo.CreateRoomType(ctx, rt)
}

func (uc *RoomUseCase) ListByHotel(ctx context.Context, hotelID string, pagination entity.PaginationRequest) ([]entity.RoomType, int64, error) {
	if hotelID == "" {
		return nil, 0, entity.ErrInvalidInput
	}
	return uc.repo.ListByHotel(ctx, hotelID, pagination)
}

func (uc *RoomUseCase) GetByID(ctx context.Context, id string) (*entity.RoomType, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, entity.ErrRecordNotFound
	}
	return uc.repo.GetByID(ctx, id)
}

func (uc *RoomUseCase) Update(ctx context.Context, id string, req entity.UpdateRoomTypeRequest) error {
	if _, err := uuid.Parse(id); err != nil {
		return entity.ErrRecordNotFound
	}
	if req.TotalQuantity != nil && *req.TotalQuantity < 0 {
		return entity.ErrInvalidInput
	}
	if req.MaxOccupancy != nil && *req.MaxOccupancy <= 0 {
		return entity.ErrInvalidInput
	}
	if req.MaxAdults != nil && *req.MaxAdults <= 0 {
		return entity.ErrInvalidInput
	}
	if req.MaxChildren != nil && *req.MaxChildren < 0 {
		return entity.ErrInvalidInput
	}
	if req.BasePrice != nil && *req.BasePrice < 0 {
		return entity.ErrInvalidInput
	}

	if req.Code != "" {
		req.Code = strings.ToUpper(req.Code)
	}
	
	return uc.repo.Update(ctx, id, req)
}

func (uc *RoomUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return entity.ErrRecordNotFound
	}
	return uc.repo.Delete(ctx, id)
}
