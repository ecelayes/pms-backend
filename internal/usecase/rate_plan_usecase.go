package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
)

type RatePlanUseCase struct {
	repo    *repository.RatePlanRepository
	resRepo *repository.ReservationRepository
}

func NewRatePlanUseCase(repo *repository.RatePlanRepository, resRepo *repository.ReservationRepository) *RatePlanUseCase {
	return &RatePlanUseCase{
		repo:    repo,
		resRepo: resRepo,
	}
}

func (uc *RatePlanUseCase) Create(ctx context.Context, req entity.CreateRatePlanRequest) (string, error) {
	if req.HotelID == "" || req.Name == "" {
		return "", errors.New("hotel_id and name are required")
	}

	if !req.CancellationPolicy.IsRefundable && len(req.CancellationPolicy.Rules) > 0 {
		return "", errors.New("non-refundable policies should not have tiered rules")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("uuid gen: %w", err)
	}

	plan := entity.RatePlan{
		BaseEntity:         entity.BaseEntity{ID: id.String()},
		HotelID:            req.HotelID,
		RoomTypeID:         req.RoomTypeID,
		Name:               req.Name,
		Description:        req.Description,
		Active:             true,
		MealPlan:           req.MealPlan,
		CancellationPolicy: req.CancellationPolicy,
		PaymentPolicy:      req.PaymentPolicy,
	}

	if err := uc.repo.Create(ctx, plan); err != nil {
		return "", err
	}

	return id.String(), nil
}

func (uc *RatePlanUseCase) ListByHotel(ctx context.Context, hotelID string) ([]entity.RatePlan, error) {
	return uc.repo.GetByHotel(ctx, hotelID)
}

func (uc *RatePlanUseCase) GetByID(ctx context.Context, id string) (*entity.RatePlan, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *RatePlanUseCase) Update(ctx context.Context, id string, req entity.UpdateRatePlanRequest) error {
	if req.CancellationPolicy != nil {
		if !req.CancellationPolicy.IsRefundable && len(req.CancellationPolicy.Rules) > 0 {
			return errors.New("cannot have cancellation rules for a non-refundable policy")
		}
	}
	
	if req.MealPlan != nil {
		if req.MealPlan.Included && req.MealPlan.PricePerPax > 0 {
			log.Printf("[WARN] RatePlan %s: MealPlan is marked 'Included' but has PricePerPax > 0. This implies a hidden surcharge.", id)
		}

		if !req.MealPlan.Included && req.MealPlan.PricePerPax == 0 {
			log.Printf("[WARN] RatePlan %s: MealPlan is 'Not Included' but PricePerPax is 0. This might mean free food not advertised.", id)
		}
	}

	return uc.repo.Update(ctx, id, req)
}

func (uc *RatePlanUseCase) Delete(ctx context.Context, id string) error {
	count, err := uc.resRepo.CountActiveByRatePlan(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check active reservations: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("cannot delete rate plan: %d active reservations depend on it", count)
	}

	return uc.repo.Delete(ctx, id)
}
