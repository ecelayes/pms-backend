package usecase

import (
	"context"
	"time"
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
	if req.Price <= 0 { return entity.ErrPriceNegative }
	if req.Priority < 0 { return entity.ErrPriorityNegative }
	
	layout := "2006-01-02"
	start, err := time.Parse(layout, req.Start)
	if err != nil { return entity.ErrInvalidDateFormat }
	end, err := time.Parse(layout, req.End)
	if err != nil { return entity.ErrInvalidDateFormat }
	if !end.After(start) { return entity.ErrInvalidDateRange }

	_, err = uc.roomRepo.GetRoomTypeByID(ctx, req.RoomTypeID)
	if err != nil { return entity.ErrRoomTypeNotFound }

	rule := entity.PriceRule{
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
