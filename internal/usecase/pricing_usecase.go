package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
	"github.com/ecelayes/pms-backend/internal/service"
)

type PricingUseCase struct {
	db             *pgxpool.Pool
	priceRepo      *repository.PriceRepository
	unitTypeRepo   *repository.UnitTypeRepository
	inventoryLogic *service.InventoryService
}

func NewPricingUseCase(
	db *pgxpool.Pool,
	priceRepo *repository.PriceRepository,
	unitTypeRepo *repository.UnitTypeRepository,
	inventoryLogic *service.InventoryService,
) *PricingUseCase {
	return &PricingUseCase{
		db:             db,
		priceRepo:      priceRepo,
		unitTypeRepo:   unitTypeRepo,
		inventoryLogic: inventoryLogic,
	}
}

func (uc *PricingUseCase) BulkCreateRule(ctx context.Context, req entity.SetPriceRequest) error {
	if _, err := uc.unitTypeRepo.GetByID(ctx, req.UnitTypeID); err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return entity.ErrUnitTypeNotFound
		}
		return err
	}

	layout := "2006-01-02"
	start, err := time.Parse(layout, req.Start)
	if err != nil { return entity.ErrInvalidInput }
	end, err := time.Parse(layout, req.End)
	if err != nil { return entity.ErrInvalidInput }

	if !end.After(start) {
		return entity.ErrInvalidInput
	}
	if req.Price < 0 {
		return entity.ErrInvalidInput
	}

	tx, err := uc.db.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)

	existingRules, err := uc.priceRepo.GetOverlapping(ctx, tx, req.UnitTypeID, start, end)
	if err != nil { return err }

	newID, _ := uuid.NewV7()
	targetRule := entity.PriceRule{
		BaseEntity: entity.BaseEntity{ID: newID.String()},
		UnitTypeID: req.UnitTypeID,
		Start:      start,
		End:        end,
		Price:      req.Price,
	}

	finalRules := uc.inventoryLogic.ResolveRuleConflicts(existingRules, targetRule)

	toDeleteIDs := []string{}
	for _, r := range existingRules {
		toDeleteIDs = append(toDeleteIDs, r.ID)
	}

	for i := range finalRules {
		if finalRules[i].ID == "" || finalRules[i].ID != targetRule.ID {
			uid, _ := uuid.NewV7()
			finalRules[i].ID = uid.String()
			finalRules[i].UnitTypeID = req.UnitTypeID
		}
	}

	if err := uc.priceRepo.BatchDelete(ctx, tx, toDeleteIDs); err != nil { return err }
	if err := uc.priceRepo.BatchCreate(ctx, tx, finalRules); err != nil { return err }

	return tx.Commit(ctx)
}

func (uc *PricingUseCase) DeleteRule(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return entity.ErrRecordNotFound
	}
	return uc.priceRepo.Delete(ctx, id)
}

func (uc *PricingUseCase) GetRules(ctx context.Context, unitTypeID, propertyID string, pagination entity.PaginationRequest) ([]entity.PriceRule, int64, error) {
	if unitTypeID != "" {
		return uc.priceRepo.ListByUnitType(ctx, unitTypeID, pagination)
	}
	if propertyID != "" {
		return uc.priceRepo.ListByProperty(ctx, propertyID, pagination)
	}
	return nil, 0, fmt.Errorf("%w: unit_type_id or property_id is required", entity.ErrInvalidInput)
}
