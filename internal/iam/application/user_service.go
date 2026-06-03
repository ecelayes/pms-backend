package application

import (
	"context"
	"errors"
	"github.com/ecelayes/pms-backend/internal/iam/domain"
	"github.com/ecelayes/pms-backend/pkg/auth"
	"strings"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserService struct {
	repo    domain.UserRepository
	orgRepo domain.OrganizationRepository
}

func NewUserService(repo domain.UserRepository, orgRepo domain.OrganizationRepository) *UserService {
	return &UserService{
		repo:    repo,
		orgRepo: orgRepo,
	}
}
func (s *UserService) Register(ctx context.Context, orgID, email, password, role, firstName, lastName, phone string) (string, error) {
	salt, err := auth.GenerateRandomSalt()
	if err != nil {
		return "", err
	}
	hashed, err := auth.HashPassword(password)
	if err != nil {
		return "", err
	}
	userRole := domain.RoleUser
	if role == "super_admin" {
		userRole = domain.RoleSuperAdmin
	}
	if role != "" {
		if role == "admin" || role == "manager" {
			userRole = domain.UserRole(role)
		}
	}
	u, err := domain.NewUser(email, hashed, salt, userRole, firstName, lastName, phone)
	if err != nil {
		return "", err
	}
	if err := s.repo.Save(ctx, u); err != nil {
		return "", err
	}
	if orgID != "" {
		if err := s.orgRepo.AddMember(ctx, orgID, u.ID(), role); err != nil {
			return "", err
		}
	}
	return u.ID(), nil
}
func (s *UserService) FindOrCreateGuest(ctx context.Context, email, firstName, lastName, phone string) (string, error) {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if u != nil {
		if err := s.repo.EnsureGuest(ctx, u); err != nil {
			return "", err
		}
		return u.ID(), nil
	}
	randomPwd, _ := auth.GenerateRandomSalt()
	id, err := s.Register(ctx, "", email, randomPwd, "user", firstName, lastName, phone)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			u, err = s.repo.FindByEmail(ctx, email)
			if err != nil {
				return "", err
			}
			if u != nil {
				if err := s.repo.EnsureGuest(ctx, u); err != nil {
					return "", err
				}
				return u.ID(), nil
			}
		}
		return "", err
	}
	u, err = s.repo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	if err := s.repo.EnsureGuest(ctx, u); err != nil {
		return "", err
	}
	return id, nil
}
func (s *UserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}
func (s *UserService) GetAll(ctx context.Context, orgID string) ([]*domain.User, error) {
	return s.repo.FindAllByOrganization(ctx, orgID)
}
func (s *UserService) Update(ctx context.Context, id string, role, firstName, lastName, phone string) error {
	u, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	u.Update(domain.UserRole(role), firstName, lastName, phone)
	return s.repo.Save(ctx, u)
}
func (s *UserService) Delete(ctx context.Context, id string) error {
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.DeleteUser(ctx, id)
}
