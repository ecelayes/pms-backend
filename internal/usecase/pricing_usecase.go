package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
)

type PricingUseCase struct {
	priceRepo *repository.PriceRepository
	roomRepo  *repository.RoomRepository
}

func NewPricingUseCase(priceRepo *repository.PriceRepository, roomRepo *repository.RoomRepository) *PricingUseCase {
	return &PricingUseCase{priceRepo: priceRepo, roomRepo: roomRepo}
}

func (uc *PricingUseCase) CreateRule(ctx context.Context, req entity.CreatePriceRuleRequest) error {
	layout := "2006-01-02"

	if req.Price <= 0 { return entity.ErrPriceNegative }
	if req.Priority < 0 { return entity.ErrPriorityNegative }

	start, err := time.Parse(layout, req.Start)
	if err != nil { return entity.ErrInvalidDateFormat }
	end, err := time.Parse(layout, req.End)
	if err != nil { return entity.ErrInvalidDateFormat }
	if !end.After(start) { return entity.ErrInvalidDateRange }

	_, err = uc.roomRepo.GetRoomTypeByID(ctx, req.RoomTypeID)
	if err != nil { return entity.ErrRoomTypeNotFound }

	newID, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("failed to generate uuid v7: %w", err)
	}

	rule := entity.PriceRule{
		BaseEntity: entity.BaseEntity{
			ID: newID.String(),
		},
		RoomTypeID: req.RoomTypeID,
		Start:      start,
		End:        end,
		Price:      req.Price,
		Priority:   req.Priority,
	}
	return uc.priceRepo.CreateRule(ctx, rule)
}

func (uc *PricingUseCase) UpdateRule(ctx context.Context, id string, req entity.UpdatePriceRuleRequest) error {
	if req.Price <= 0 { return entity.ErrPriceNegative }
	if req.Priority < 0 { return entity.ErrPriorityNegative }
	return uc.priceRepo.Update(ctx, id, req)
}

func (uc *PricingUseCase) DeleteRule(ctx context.Context, id string) error {
	return uc.priceRepo.Delete(ctx, id)
}
