package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
)

type CatalogUseCase struct {
	amenityRepo *repository.AmenityRepository
	serviceRepo *repository.HotelServiceRepository
}

func NewCatalogUseCase(ar *repository.AmenityRepository, sr *repository.HotelServiceRepository) *CatalogUseCase {
	return &CatalogUseCase{amenityRepo: ar, serviceRepo: sr}
}

func (uc *CatalogUseCase) CreateAmenity(ctx context.Context, requesterRole string, req entity.CreateCatalogRequest) (string, error) {
	if requesterRole != entity.RoleSuperAdmin {
		return "", errors.New("insufficient permissions: only super admin can create global catalogs")
	}

	id, _ := uuid.NewV7()
	a := entity.Amenity{
		BaseEntity:  entity.BaseEntity{ID: id.String()},
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
	}

	if err := uc.amenityRepo.Create(ctx, a); err != nil {
		return "", err
	}
	return id.String(), nil
}

func (uc *CatalogUseCase) GetAllAmenities(ctx context.Context) ([]entity.Amenity, error) {
	return uc.amenityRepo.GetAll(ctx)
}

func (uc *CatalogUseCase) GetAmenityByID(ctx context.Context, id string) (*entity.Amenity, error) {
	return uc.amenityRepo.GetByID(ctx, id)
}

func (uc *CatalogUseCase) UpdateAmenity(ctx context.Context, role, id string, req entity.UpdateCatalogRequest) error {
	if role != entity.RoleSuperAdmin {
		return entity.ErrInsufficientPermissions
	}
	return uc.amenityRepo.Update(ctx, id, req)
}

func (uc *CatalogUseCase) DeleteAmenity(ctx context.Context, role, id string) error {
	if role != entity.RoleSuperAdmin {
		return entity.ErrInsufficientPermissions
	}
	return uc.amenityRepo.Delete(ctx, id)
}

func (uc *CatalogUseCase) CreateService(ctx context.Context, requesterRole string, req entity.CreateCatalogRequest) (string, error) {
	if requesterRole != entity.RoleSuperAdmin {
		return "", errors.New("insufficient permissions")
	}
	
	id, _ := uuid.NewV7()
	s := entity.HotelService{
		BaseEntity:  entity.BaseEntity{ID: id.String()},
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
	}

	if err := uc.serviceRepo.Create(ctx, s); err != nil {
		return "", err
	}
	return id.String(), nil
}

func (uc *CatalogUseCase) GetAllServices(ctx context.Context) ([]entity.HotelService, error) {
	return uc.serviceRepo.GetAll(ctx)
}

func (uc *CatalogUseCase) GetServiceByID(ctx context.Context, id string) (*entity.HotelService, error) {
	return uc.serviceRepo.GetByID(ctx, id)
}

func (uc *CatalogUseCase) UpdateService(ctx context.Context, role, id string, req entity.UpdateCatalogRequest) error {
	if role != entity.RoleSuperAdmin {
		return entity.ErrInsufficientPermissions
	}
	return uc.serviceRepo.Update(ctx, id, req)
}

func (uc *CatalogUseCase) DeleteService(ctx context.Context, role, id string) error {
	if role != entity.RoleSuperAdmin {
		return entity.ErrInsufficientPermissions
	}
	return uc.serviceRepo.Delete(ctx, id)
}
