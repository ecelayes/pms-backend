package usecase

import (
	"context"
	"errors"
	"strings"

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
		return "", errors.New("hotel_id and name are required")
	}
	if len(req.Code) != 3 {
		return "", errors.New("room code must be exactly 3 characters")
	}
	if req.TotalQuantity < 0 {
		return "", errors.New("total quantity cannot be negative")
	}
	if req.MaxOccupancy <= 0 {
		return "", errors.New("max occupancy must be at least 1")
	}
	if req.MaxAdults <= 0 {
		return "", errors.New("max adults must be at least 1")
	}

	rt := entity.RoomType{
		HotelID:       req.HotelID,
		Name:          req.Name,
		Code:          strings.ToUpper(req.Code),
		TotalQuantity: req.TotalQuantity,
		
		MaxOccupancy:  req.MaxOccupancy,
		MaxAdults:     req.MaxAdults,
		MaxChildren:   req.MaxChildren,
		Amenities:     req.Amenities,
	}

	return uc.repo.CreateRoomType(ctx, rt)
}

func (uc *RoomUseCase) Update(ctx context.Context, id string, req entity.UpdateRoomTypeRequest) error {
	if req.TotalQuantity < 0 {
		return errors.New("total quantity cannot be negative")
	}
	if req.MaxOccupancy <= 0 {
		return errors.New("max occupancy must be at least 1")
	}
	if req.MaxAdults <= 0 {
		return errors.New("max adults must be at least 1")
	}
	if req.Code != "" {
		req.Code = strings.ToUpper(req.Code)
	}
	return uc.repo.Update(ctx, id, req)
}

func (uc *RoomUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}
