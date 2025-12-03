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

type CreateRoomTypeRequest struct {
	HotelID       string `json:"hotel_id"`
	Name          string `json:"name"`
	Code          string `json:"code"`
	TotalQuantity int    `json:"total_quantity"`
}

func (uc *RoomUseCase) Create(ctx context.Context, req CreateRoomTypeRequest) (string, error) {
	if req.HotelID == "" || req.Name == "" {
		return "", errors.New("hotel_id and name are required")
	}
	if len(req.Code) != 3 {
		return "", errors.New("room code must be exactly 3 characters")
	}
	if req.TotalQuantity < 0 {
		return "", errors.New("total quantity cannot be negative")
	}

	code := strings.ToUpper(req.Code)

	rt := entity.RoomType{
		HotelID:       req.HotelID,
		Name:          req.Name,
		Code:          code,
		TotalQuantity: req.TotalQuantity,
	}

	return uc.repo.CreateRoomType(ctx, rt)
}
