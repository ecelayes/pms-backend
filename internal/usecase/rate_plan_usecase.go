package usecase

import (
	"context"
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
	if req.PropertyID == "" || req.Name == "" {
		return "", entity.ErrInvalidInput
	}

	if !req.CancellationPolicy.IsRefundable && len(req.CancellationPolicy.Rules) > 0 {
		return "", entity.ErrInvalidInput
	}

	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("uuid gen: %w", err)
	}

	plan := entity.RatePlan{
		BaseEntity:         entity.BaseEntity{ID: id.String()},
		PropertyID:         req.PropertyID,
		UnitTypeID:         req.UnitTypeID,
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

func (uc *RatePlanUseCase) ListByProperty(ctx context.Context, propertyID string, pagination entity.PaginationRequest) ([]entity.RatePlan, int64, error) {
	return uc.repo.ListByProperty(ctx, propertyID, pagination)
}

func (uc *RatePlanUseCase) GetByID(ctx context.Context, id string) (*entity.RatePlan, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, entity.ErrRecordNotFound
	}
	return uc.repo.GetByID(ctx, id)
}

func (uc *RatePlanUseCase) Update(ctx context.Context, id string, req entity.UpdateRatePlanRequest) error {
	if _, err := uuid.Parse(id); err != nil {
		return entity.ErrRecordNotFound
	}

	if req.CancellationPolicy != nil {
		if !req.CancellationPolicy.IsRefundable && len(req.CancellationPolicy.Rules) > 0 {
			return entity.ErrInvalidInput
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
	if _, err := uuid.Parse(id); err != nil {
		return entity.ErrRecordNotFound
	}

	count, err := uc.resRepo.CountActiveByRatePlan(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check active reservations: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("cannot delete rate plan: %d active reservations depend on it: %w", count, entity.ErrConflict)
	}

	return uc.repo.Delete(ctx, id)
}
