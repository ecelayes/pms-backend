package application

import (
	"context"
	"errors"
	"github.com/ecelayes/pms-backend/internal/catalog/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
)

var (
	ErrNotFound = domain.ErrNotFound
)

type CatalogService struct {
	unitTypeRepo     domain.UnitTypeRepository
	propRepo         domain.PropertyRepository
	unitRepo         domain.UnitRepository
	amenityRepo      domain.AmenityRepository
	guestServiceRepo domain.GuestServiceRepository
}

func NewCatalogService(
	utr domain.UnitTypeRepository,
	pr domain.PropertyRepository,
	ur domain.UnitRepository,
	ar domain.AmenityRepository,
	hsr domain.GuestServiceRepository,
) *CatalogService {
	return &CatalogService{
		unitTypeRepo:     utr,
		propRepo:         pr,
		unitRepo:         ur,
		amenityRepo:      ar,
		guestServiceRepo: hsr,
	}
}
func (s *CatalogService) CreateProperty(ctx context.Context, orgID, name, code, pTypeStr string) (string, error) {
	pType := domain.PropertyType(pTypeStr)
	if pType == "" {
		pType = domain.PropertyTypeHotel
	}
	prop, err := domain.NewProperty(orgID, name, code, pType)
	if err != nil {
		return "", err
	}
	if err := s.propRepo.SaveProperty(ctx, prop); err != nil {
		return "", err
	}
	return prop.ID(), nil
}
func (s *CatalogService) GetProperty(ctx context.Context, id string) (*domain.Property, error) {
	p, err := s.propRepo.FindPropertyByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrNotFound
	}
	return p, nil
}
func (s *CatalogService) ListProperties(ctx context.Context, orgID string, page, limit int) ([]*domain.Property, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	return s.propRepo.FindPropertiesByOrganization(ctx, orgID, limit, offset)
}
func (s *CatalogService) UpdateProperty(ctx context.Context, id, name, code, pTypeStr string) error {
	p, err := s.GetProperty(ctx, id)
	if err != nil {
		return err
	}
	p.Update(name, code, domain.PropertyType(pTypeStr))
	return s.propRepo.SaveProperty(ctx, p)
}
func (s *CatalogService) DeleteProperty(ctx context.Context, id string) error {
	return s.propRepo.DeleteProperty(ctx, id)
}
func (s *CatalogService) GetUnitType(ctx context.Context, id string) (string, vo.Money, int, error) {
	ut, err := s.unitTypeRepo.FindUnitTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", vo.Money{}, 0, ErrNotFound
		}
		return "", vo.Money{}, 0, err
	}
	if ut == nil {
		return "", vo.Money{}, 0, ErrNotFound
	}
	return ut.PropertyID(), ut.BasePrice(), ut.TotalQuantity(), nil
}
func (s *CatalogService) GetUnitTypeEntity(ctx context.Context, id string) (*domain.UnitType, error) {
	ut, err := s.unitTypeRepo.FindUnitTypeByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ut == nil {
		return nil, ErrNotFound
	}
	return ut, nil
}
func (s *CatalogService) CreateUnitType(ctx context.Context, propertyID, name, code string, qty int, priceCents int64, currency string, maxOcc, maxAd, maxCh int, amenities []string) (string, error) {
	price := vo.NewMoney(priceCents, currency)
	ut, err := domain.NewUnitType(propertyID, name, code, qty, price, maxOcc, maxAd, maxCh, amenities)
	if err != nil {
		return "", err
	}
	if err := s.unitTypeRepo.SaveUnitType(ctx, ut); err != nil {
		return "", err
	}
	return ut.ID(), nil
}
func (s *CatalogService) ListUnitTypes(ctx context.Context, propertyID string, page, limit int) ([]*domain.UnitType, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	return s.unitTypeRepo.FindUnitTypesByPropertyID(ctx, propertyID, limit, offset)
}
func (s *CatalogService) UpdateUnitType(ctx context.Context, id, name, code string, qty int, priceCents int64, currency string, maxOcc, maxAd, maxCh int, amenities []string) error {
	ut, err := s.GetUnitTypeEntity(ctx, id)
	if err != nil {
		return err
	}
	price := vo.NewMoney(priceCents, currency)
	ut.Update(name, code, qty, price, maxOcc, maxAd, maxCh, amenities)
	return s.unitTypeRepo.SaveUnitType(ctx, ut)
}
func (s *CatalogService) DeleteUnitType(ctx context.Context, id string) error {
	return s.unitTypeRepo.DeleteUnitType(ctx, id)
}
func (s *CatalogService) CreateUnit(ctx context.Context, propertyID, unitTypeID, name string) (string, error) {
	u := domain.NewUnit(propertyID, unitTypeID, name)
	if err := s.unitRepo.SaveUnit(ctx, u); err != nil {
		return "", err
	}
	return u.ID(), nil
}
func (s *CatalogService) GetUnit(ctx context.Context, id string) (*domain.Unit, error) {
	u, err := s.unitRepo.FindUnitByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return u, nil
}
func (s *CatalogService) ListUnits(ctx context.Context, propertyID string, page, limit int) ([]*domain.Unit, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	return s.unitRepo.FindUnitsByPropertyID(ctx, propertyID, limit, offset)
}
func (s *CatalogService) UpdateUnit(ctx context.Context, id, name, status string) error {
	u, err := s.GetUnit(ctx, id)
	if err != nil {
		return err
	}
	u.Update(name, domain.UnitStatus(status))
	return s.unitRepo.SaveUnit(ctx, u)
}
func (s *CatalogService) DeleteUnit(ctx context.Context, id string) error {
	return s.unitRepo.DeleteUnit(ctx, id)
}
func (s *CatalogService) CreateAmenity(ctx context.Context, name, description, icon string) (string, error) {
	a, err := domain.NewAmenity(name, description, icon)
	if err != nil {
		return "", err
	}
	if err := s.amenityRepo.Save(ctx, a); err != nil {
		return "", err
	}
	return a.ID(), nil
}
func (s *CatalogService) GetAmenity(ctx context.Context, id string) (*domain.Amenity, error) {
	a, err := s.amenityRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, ErrNotFound
	}
	return a, nil
}
func (s *CatalogService) ListAmenities(ctx context.Context, page, limit int) ([]*domain.Amenity, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	return s.amenityRepo.FindAll(ctx, limit, offset)
}
func (s *CatalogService) UpdateAmenity(ctx context.Context, id, name, description, icon string) error {
	a, err := s.GetAmenity(ctx, id)
	if err != nil {
		return err
	}
	a.Update(name, description, icon)
	return s.amenityRepo.Save(ctx, a)
}
func (s *CatalogService) DeleteAmenity(ctx context.Context, id string) error {
	return s.amenityRepo.Delete(ctx, id)
}
func (s *CatalogService) CreateGuestService(ctx context.Context, name, description, icon string) (string, error) {
	gs, err := domain.NewGuestService(name, description, icon)
	if err != nil {
		return "", err
	}
	if err := s.guestServiceRepo.SaveGuestService(ctx, gs); err != nil {
		return "", err
	}
	return gs.ID(), nil
}
func (s *CatalogService) GetGuestService(ctx context.Context, id string) (*domain.GuestService, error) {
	gs, err := s.guestServiceRepo.FindGuestServiceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if gs == nil {
		return nil, ErrNotFound
	}
	return gs, nil
}
func (s *CatalogService) ListGuestServices(ctx context.Context, page, limit int) ([]*domain.GuestService, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	return s.guestServiceRepo.FindAllGuestServices(ctx, limit, offset)
}
func (s *CatalogService) UpdateGuestService(ctx context.Context, id, name, description, icon string) error {
	gs, err := s.GetGuestService(ctx, id)
	if err != nil {
		return err
	}
	gs.Update(name, description, icon)
	return s.guestServiceRepo.SaveGuestService(ctx, gs)
}
func (s *CatalogService) DeleteGuestService(ctx context.Context, id string) error {
	return s.guestServiceRepo.DeleteGuestService(ctx, id)
}
