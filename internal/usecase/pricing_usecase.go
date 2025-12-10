package usecase

import (
	"context"
	"errors"
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
	return &PricingUseCase{
		priceRepo: priceRepo,
		roomRepo:  roomRepo,
	}
}

func (uc *PricingUseCase) CreateRule(ctx context.Context, req entity.CreatePriceRuleRequest) error {
	if _, err := uc.roomRepo.GetByID(ctx, req.RoomTypeID); err != nil {
		return errors.New("room type not found")
	}

	layout := "2006-01-02"
	start, err := time.Parse(layout, req.Start)
	if err != nil {
		return errors.New("invalid start date format (YYYY-MM-DD)")
	}
	end, err := time.Parse(layout, req.End)
	if err != nil {
		return errors.New("invalid end date format (YYYY-MM-DD)")
	}

	if !end.After(start) {
		return errors.New("end date must be after start date")
	}

	priceID, err := uuid.NewV7()
	if err != nil {
		return err
	}

	rule := entity.PriceRule{
		BaseEntity: entity.BaseEntity{
			ID: priceID.String(),
		},
		RoomTypeID: req.RoomTypeID,
		Start:      start,
		End:        end,
		Price:      req.Price,
		Priority:   req.Priority,
	}

	return uc.priceRepo.Create(ctx, rule)
}

func (uc *PricingUseCase) GetRules(ctx context.Context, roomTypeID string) ([]entity.PriceRule, error) {
	if roomTypeID == "" {
		return nil, errors.New("room_type_id is required")
	}
	return uc.priceRepo.ListByRoomType(ctx, roomTypeID)
}

func (uc *PricingUseCase) UpdateRule(ctx context.Context, id string, req entity.UpdatePriceRuleRequest) error {
	if req.Price < 0 {
		return errors.New("price cannot be negative")
	}
	return uc.priceRepo.Update(ctx, id, req)
}

func (uc *PricingUseCase) DeleteRule(ctx context.Context, id string) error {
	rule, err := uc.priceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()

	isActive := (now.Equal(rule.Start) || now.After(rule.Start)) && now.Before(rule.End)

	if isActive {
		return errors.New("cannot delete an active price rule (currently in effect)")
	}

	return uc.priceRepo.Delete(ctx, id)
}
