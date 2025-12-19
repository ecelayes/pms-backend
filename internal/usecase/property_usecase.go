package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
)

type PropertyUseCase struct {
	repo *repository.PropertyRepository
}

func NewPropertyUseCase(repo *repository.PropertyRepository) *PropertyUseCase {
	return &PropertyUseCase{repo: repo}
}

func (uc *PropertyUseCase) Create(ctx context.Context, req entity.CreatePropertyRequest) (string, error) {
	if req.OrganizationID == "" {
		return "", entity.ErrInvalidInput
	}
	if req.Name == "" || req.Code == "" {
		return "", entity.ErrInvalidInput
	}
	if len(req.Code) < 3 || len(req.Code) > 5 {
		return "", entity.ErrInvalidInput
	}
	
	propertyID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid v7: %w", err)
	}

	property := entity.Property{
		BaseEntity: entity.BaseEntity{
			ID: propertyID.String(),
		},
		OrganizationID: req.OrganizationID, 
		Name:           req.Name,
		Code:           strings.ToUpper(req.Code),
		Type:           req.Type,
	}
	if property.Type == "" {
		property.Type = "HOTEL"
	}

	return uc.repo.Create(ctx, property)
}

func (uc *PropertyUseCase) ListByOrganization(ctx context.Context, orgID string, pagination entity.PaginationRequest) ([]entity.Property, int64, error) {
	if orgID == "" {
		return nil, 0, entity.ErrInvalidInput
	}
	return uc.repo.ListByOrganization(ctx, orgID, pagination)
}

func (uc *PropertyUseCase) GetByID(ctx context.Context, id string) (*entity.Property, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, entity.ErrRecordNotFound
	}
	return uc.repo.GetByID(ctx, id)
}

func (uc *PropertyUseCase) Update(ctx context.Context, id string, req entity.UpdatePropertyRequest) error {
	if _, err := uuid.Parse(id); err != nil {
		return entity.ErrRecordNotFound
	}
	if req.Name == "" && req.Code == "" && req.Type == "" {
		return entity.ErrInvalidInput
	}
	if req.Code != "" {
		req.Code = strings.ToUpper(req.Code)
	}
	return uc.repo.Update(ctx, id, req)
}

func (uc *PropertyUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return entity.ErrRecordNotFound
	}
	return uc.repo.Delete(ctx, id)
}
