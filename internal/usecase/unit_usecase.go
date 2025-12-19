package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
)

type UnitUseCase struct {
	repo *repository.UnitRepository
}

func NewUnitUseCase(repo *repository.UnitRepository) *UnitUseCase {
	return &UnitUseCase{repo: repo}
}

func (uc *UnitUseCase) Create(ctx context.Context, req entity.CreateUnitRequest) (string, error) {
	if req.PropertyID == "" || req.UnitTypeID == "" || req.Name == "" {
		return "", entity.ErrInvalidInput
	}

	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("uuid gen: %w", err)
	}

	unit := entity.Unit{
		BaseEntity: entity.BaseEntity{ID: id.String()},
		PropertyID: req.PropertyID,
		UnitTypeID: req.UnitTypeID,
		Name:       req.Name,
		Status:     req.Status,
	}

	if unit.Status == "" {
		unit.Status = "CLEAN"
	}

	if err := uc.repo.Create(ctx, unit); err != nil {
		return "", err
	}

	return id.String(), nil
}

func (uc *UnitUseCase) ListByProperty(ctx context.Context, propertyID string) ([]entity.Unit, error) {
	return uc.repo.ListByProperty(ctx, propertyID)
}

func (uc *UnitUseCase) GetByID(ctx context.Context, id string) (*entity.Unit, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, entity.ErrRecordNotFound
	}
	return uc.repo.GetByID(ctx, id)
}

func (uc *UnitUseCase) Update(ctx context.Context, id string, req entity.UpdateUnitRequest) error {
	if _, err := uuid.Parse(id); err != nil {
		return entity.ErrRecordNotFound
	}
	return uc.repo.Update(ctx, id, req)
}

func (uc *UnitUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return entity.ErrRecordNotFound
	}
	return uc.repo.Delete(ctx, id)
}
