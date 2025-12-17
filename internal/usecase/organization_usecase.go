package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
	"github.com/ecelayes/pms-backend/internal/utils"
)

type OrganizationUseCase struct {
	repo *repository.OrganizationRepository
}

func NewOrganizationUseCase(repo *repository.OrganizationRepository) *OrganizationUseCase {
	return &OrganizationUseCase{repo: repo}
}

func (uc *OrganizationUseCase) Create(ctx context.Context, req entity.CreateOrganizationRequest) (string, error) {
	if req.Name == "" {
		return "", entity.ErrInvalidInput
	}

	orgID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid: %w", err)
	}

	slug := strings.ReplaceAll(strings.ToUpper(req.Name), " ", "")
	if len(slug) > 5 {
		slug = slug[:5]
	}
	code := fmt.Sprintf("%s-%s", slug, utils.GenerateRandomCode(3))

	org := entity.Organization{
		BaseEntity: entity.BaseEntity{ID: orgID.String()},
		Name:       req.Name,
		Code:       code,
	}

	if err := uc.repo.Create(ctx, nil, org); err != nil {
		return "", err
	}

	return orgID.String(), nil
}

func (uc *OrganizationUseCase) GetAll(ctx context.Context, pagination entity.PaginationRequest) ([]entity.Organization, int64, error) {
	return uc.repo.GetAll(ctx, pagination)
}

func (uc *OrganizationUseCase) GetByID(ctx context.Context, id string) (*entity.Organization, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, entity.ErrRecordNotFound
	}
	return uc.repo.GetByID(ctx, id)
}

func (uc *OrganizationUseCase) Update(ctx context.Context, id string, req entity.UpdateOrganizationRequest) error {
	if _, err := uuid.Parse(id); err != nil {
		return entity.ErrRecordNotFound
	}
	if req.Name == "" {
		return entity.ErrInvalidInput
	}
	return uc.repo.Update(ctx, id, req)
}

func (uc *OrganizationUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return entity.ErrRecordNotFound
	}
	return uc.repo.Delete(ctx, id)
}
