package application

import (
	"context"
	"errors"
	"github.com/ecelayes/pms-backend/internal/iam/domain"
)

var (
	ErrOrgNotFound = errors.New("organization not found")
)

type OrganizationService struct {
	repo domain.OrganizationRepository
}

func NewOrganizationService(repo domain.OrganizationRepository) *OrganizationService {
	return &OrganizationService{repo: repo}
}
func (s *OrganizationService) Create(ctx context.Context, name, code string) (string, error) {
	org, err := domain.NewOrganization(name, code)
	if err != nil {
		return "", err
	}
	if err := s.repo.Save(ctx, org); err != nil {
		return "", err
	}
	return org.ID(), nil
}
func (s *OrganizationService) GetByID(ctx context.Context, id string) (*domain.Organization, error) {
	org, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, ErrOrgNotFound
	}
	return org, nil
}
func (s *OrganizationService) GetAll(ctx context.Context) ([]*domain.Organization, error) {
	return s.repo.FindAll(ctx)
}
func (s *OrganizationService) Update(ctx context.Context, id, name, code string) error {
	org, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	org.Update(name, code)
	return s.repo.Save(ctx, org)
}
func (s *OrganizationService) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOrganization(ctx, id)
}
