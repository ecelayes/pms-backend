package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
	"github.com/ecelayes/pms-backend/pkg/auth"
)

type UserUseCase struct {
	db       *pgxpool.Pool
	userRepo *repository.UserRepository
	orgRepo  *repository.OrganizationRepository
}

func NewUserUseCase(db *pgxpool.Pool, userRepo *repository.UserRepository, orgRepo *repository.OrganizationRepository) *UserUseCase {
	return &UserUseCase{
		db:       db,
		userRepo: userRepo,
		orgRepo:  orgRepo,
	}
}

func (uc *UserUseCase) CreateUser(ctx context.Context, requesterRole string, req entity.CreateUserRequest) (string, error) {
	if req.Email == "" || req.Password == "" {
		return "", errors.New("email and password are required")
	}
	if req.OrganizationID == "" {
		return "", errors.New("organization_id is required")
	}

	switch req.Role {
	case entity.OrgRoleOwner:
		if requesterRole != entity.RoleSuperAdmin {
			return "", errors.New("only super_admin can create organization owners")
		}
	
	case entity.OrgRoleManager, entity.OrgRoleStaff:
		if requesterRole != entity.RoleSuperAdmin && requesterRole != entity.OrgRoleOwner {
			return "", errors.New("insufficient permissions to create staff")
		}
	
	default:
		return "", errors.New("invalid target role")
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil { return "", err }
	
	userSalt, err := auth.GenerateRandomSalt()
	if err != nil { return "", err }

	userID, err := uuid.NewV7()
	if err != nil { return "", err }
	
	memberID, err := uuid.NewV7()
	if err != nil { return "", err }

	tx, err := uc.db.Begin(ctx)
	if err != nil { return "", err }
	defer tx.Rollback(ctx)

	err = uc.userRepo.Create(ctx, tx, userID.String(), req.Email, passwordHash, userSalt, entity.RoleUser)
	if err != nil { return "", err }

	member := entity.OrganizationMember{
		BaseEntity:     entity.BaseEntity{ID: memberID.String()},
		OrganizationID: req.OrganizationID,
		UserID:         userID.String(),
		Role:           req.Role,
	}
	err = uc.orgRepo.AddMember(ctx, tx, member)
	if err != nil { return "", err }

	if err := tx.Commit(ctx); err != nil { return "", err }

	return userID.String(), nil
}

func (uc *UserUseCase) GetAll(ctx context.Context, orgID string) ([]entity.User, error) {
	if orgID == "" {
		return nil, errors.New("organization_id is required")
	}
	return uc.userRepo.GetAllByOrganization(ctx, orgID)
}

func (uc *UserUseCase) GetByID(ctx context.Context, id string) (*entity.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}

func (uc *UserUseCase) Update(ctx context.Context, id string, orgID string, req entity.UpdateUserRequest) error {
	if req.Role != "" {
		switch req.Role {
		case entity.OrgRoleOwner, entity.OrgRoleManager, entity.OrgRoleStaff:
		default:
			return errors.New("invalid role: must be 'owner', 'manager' or 'staff'")
		}
	}

	return uc.userRepo.Update(ctx, id, orgID, req)
}

func (uc *UserUseCase) Delete(ctx context.Context, id string) error {
	return uc.userRepo.Delete(ctx, id)
}
