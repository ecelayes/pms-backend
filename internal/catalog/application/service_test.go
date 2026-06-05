package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ecelayes/pms-backend/internal/catalog/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
)

type mockPropertyRepo struct {
	prop           *domain.Property
	props          []*domain.Property
	saveErr        error
	findErr        error
	deleteErr      error
	countResult    int64
}

func (m *mockPropertyRepo) SaveProperty(ctx context.Context, p *domain.Property) error {
	return m.saveErr
}
func (m *mockPropertyRepo) FindPropertyByID(ctx context.Context, id string) (*domain.Property, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.prop, nil
}
func (m *mockPropertyRepo) FindPropertiesByOrganization(ctx context.Context, orgID string, limit, offset int) ([]*domain.Property, int64, error) {
	return m.props, int64(len(m.props)), nil
}
func (m *mockPropertyRepo) DeleteProperty(ctx context.Context, id string) error {
	return m.deleteErr
}

type mockUnitTypeRepo struct {
	unitType  *domain.UnitType
	unitTypes []*domain.UnitType
	saveErr   error
	findErr   error
	deleteErr error
}

func (m *mockUnitTypeRepo) SaveUnitType(ctx context.Context, ut *domain.UnitType) error {
	return m.saveErr
}
func (m *mockUnitTypeRepo) FindUnitTypeByID(ctx context.Context, id string) (*domain.UnitType, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.unitType, nil
}
func (m *mockUnitTypeRepo) FindUnitTypesByPropertyID(ctx context.Context, propertyID string, limit, offset int) ([]*domain.UnitType, int64, error) {
	return m.unitTypes, int64(len(m.unitTypes)), nil
}
func (m *mockUnitTypeRepo) DeleteUnitType(ctx context.Context, id string) error {
	return m.deleteErr
}

type mockUnitRepo struct {
	unit   *domain.Unit
	units  []*domain.Unit
	saveErr error
	findErr  error
	deleteErr error
}

func (m *mockUnitRepo) SaveUnit(ctx context.Context, u *domain.Unit) error {
	return m.saveErr
}
func (m *mockUnitRepo) FindUnitByID(ctx context.Context, id string) (*domain.Unit, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.unit, nil
}
func (m *mockUnitRepo) FindUnitsByPropertyID(ctx context.Context, propertyID string, limit, offset int) ([]*domain.Unit, int64, error) {
	return m.units, int64(len(m.units)), nil
}
func (m *mockUnitRepo) DeleteUnit(ctx context.Context, id string) error {
	return m.deleteErr
}

type mockAmenityRepo struct {
	amenity  *domain.Amenity
	amenities []*domain.Amenity
	saveErr   error
	findErr   error
	deleteErr error
}

func (m *mockAmenityRepo) Save(ctx context.Context, a *domain.Amenity) error {
	return m.saveErr
}
func (m *mockAmenityRepo) FindByID(ctx context.Context, id string) (*domain.Amenity, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.amenity, nil
}
func (m *mockAmenityRepo) FindAll(ctx context.Context, limit, offset int) ([]*domain.Amenity, int64, error) {
	return m.amenities, int64(len(m.amenities)), nil
}
func (m *mockAmenityRepo) Delete(ctx context.Context, id string) error {
	return m.deleteErr
}

type mockGuestServiceRepo struct {
	guestService     *domain.GuestService
	guestServices    []*domain.GuestService
	saveErr           error
	findErr           error
	deleteErr         error
}

func (m *mockGuestServiceRepo) SaveGuestService(ctx context.Context, gs *domain.GuestService) error {
	return m.saveErr
}
func (m *mockGuestServiceRepo) FindGuestServiceByID(ctx context.Context, id string) (*domain.GuestService, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.guestService, nil
}
func (m *mockGuestServiceRepo) FindAllGuestServices(ctx context.Context, limit, offset int) ([]*domain.GuestService, int64, error) {
	return m.guestServices, int64(len(m.guestServices)), nil
}
func (m *mockGuestServiceRepo) DeleteGuestService(ctx context.Context, id string) error {
	return m.deleteErr
}

func TestCatalogService_CreateProperty_Success(t *testing.T) {
	repo := &mockPropertyRepo{}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	id, err := svc.CreateProperty(context.Background(), "org-1", "Test Hotel", "THH", "HOTEL")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id == "" {
		t.Error("Expected non-empty id")
	}
}

func TestCatalogService_CreateProperty_SaveError(t *testing.T) {
	repo := &mockPropertyRepo{saveErr: errors.New("save error")}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	_, err := svc.CreateProperty(context.Background(), "org-1", "Test Hotel", "THH", "HOTEL")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestCatalogService_GetProperty_Success(t *testing.T) {
	prop, _ := domain.NewProperty("org-1", "Test Hotel", "THH", domain.PropertyTypeHotel)
	repo := &mockPropertyRepo{prop: prop}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	result, err := svc.GetProperty(context.Background(), prop.ID())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Name() != "Test Hotel" {
		t.Errorf("Expected 'Test Hotel', got '%s'", result.Name())
	}
}

func TestCatalogService_GetProperty_NotFound(t *testing.T) {
	repo := &mockPropertyRepo{prop: nil}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	_, err := svc.GetProperty(context.Background(), "non-existent")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCatalogService_GetProperty_RepoError(t *testing.T) {
	repo := &mockPropertyRepo{findErr: errors.New("repo error")}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	_, err := svc.GetProperty(context.Background(), "id")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestCatalogService_ListProperties_Success(t *testing.T) {
	prop, _ := domain.NewProperty("org-1", "Test Hotel", "THH", domain.PropertyTypeHotel)
	repo := &mockPropertyRepo{props: []*domain.Property{prop}}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	props, count, err := svc.ListProperties(context.Background(), "org-1", 1, 10)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(props) != 1 {
		t.Errorf("Expected 1 property, got %d", len(props))
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestCatalogService_ListProperties_DefaultPagination(t *testing.T) {
	repo := &mockPropertyRepo{props: []*domain.Property{}}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	_, _, err := svc.ListProperties(context.Background(), "org-1", 0, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_UpdateProperty_Success(t *testing.T) {
	prop, _ := domain.NewProperty("org-1", "Test Hotel", "THH", domain.PropertyTypeHotel)
	repo := &mockPropertyRepo{prop: prop}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	err := svc.UpdateProperty(context.Background(), prop.ID(), "Updated Hotel", "UHH", "HOTEL")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_UpdateProperty_NotFound(t *testing.T) {
	repo := &mockPropertyRepo{prop: nil}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	err := svc.UpdateProperty(context.Background(), "non-existent", "Updated", "UHH", "HOTEL")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCatalogService_DeleteProperty_Success(t *testing.T) {
	repo := &mockPropertyRepo{}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	err := svc.DeleteProperty(context.Background(), "prop-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_GetUnitType_Success(t *testing.T) {
	ut, _ := domain.NewUnitType("prop-1", "Standard", "STD", 10, vo.NewMoney(15000, "USD"), 2, 2, 0, nil)
	repo := &mockUnitTypeRepo{unitType: ut}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	propID, _, qty, err := svc.GetUnitType(context.Background(), ut.ID())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if propID != "prop-1" {
		t.Errorf("Expected prop-1, got %s", propID)
	}
	if qty != 10 {
		t.Errorf("Expected qty 10, got %d", qty)
	}
}

func TestCatalogService_GetUnitType_NotFound(t *testing.T) {
	repo := &mockUnitTypeRepo{unitType: nil}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	_, _, _, err := svc.GetUnitType(context.Background(), "non-existent")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCatalogService_CreateUnitType_Success(t *testing.T) {
	repo := &mockUnitTypeRepo{}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	id, err := svc.CreateUnitType(context.Background(), "prop-1", "Deluxe", "DLX", 5, 19999, "USD", 2, 2, 1, []string{"wifi"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id == "" {
		t.Error("Expected non-empty id")
	}
}

func TestCatalogService_CreateUnitType_SaveError(t *testing.T) {
	repo := &mockUnitTypeRepo{saveErr: errors.New("save error")}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	_, err := svc.CreateUnitType(context.Background(), "prop-1", "Deluxe", "DLX", 5, 19999, "USD", 2, 2, 1, nil)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestCatalogService_ListUnitTypes_Success(t *testing.T) {
	ut, _ := domain.NewUnitType("prop-1", "Standard", "STD", 10, vo.NewMoney(15000, "USD"), 2, 2, 0, nil)
	repo := &mockUnitTypeRepo{unitTypes: []*domain.UnitType{ut}}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	types, count, err := svc.ListUnitTypes(context.Background(), "prop-1", 1, 10)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(types) != 1 {
		t.Errorf("Expected 1 unit type, got %d", len(types))
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestCatalogService_ListUnitTypes_DefaultPagination(t *testing.T) {
	repo := &mockUnitTypeRepo{unitTypes: []*domain.UnitType{}}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	_, _, err := svc.ListUnitTypes(context.Background(), "prop-1", 0, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_UpdateUnitType_Success(t *testing.T) {
	ut, _ := domain.NewUnitType("prop-1", "Standard", "STD", 10, vo.NewMoney(15000, "USD"), 2, 2, 0, nil)
	repo := &mockUnitTypeRepo{unitType: ut}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	err := svc.UpdateUnitType(context.Background(), ut.ID(), "Deluxe", "DLX", 5, 19999, "USD", 2, 2, 1, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_UpdateUnitType_NotFound(t *testing.T) {
	repo := &mockUnitTypeRepo{unitType: nil}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	err := svc.UpdateUnitType(context.Background(), "non-existent", "Deluxe", "DLX", 5, 19999, "USD", 2, 2, 1, nil)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCatalogService_DeleteUnitType_Success(t *testing.T) {
	repo := &mockUnitTypeRepo{}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	err := svc.DeleteUnitType(context.Background(), "ut-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_CreateUnit_Success(t *testing.T) {
	repo := &mockUnitRepo{}
	svc := NewCatalogService(nil, nil, repo, nil, nil)

	id, err := svc.CreateUnit(context.Background(), "prop-1", "ut-1", "Room 101")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id == "" {
		t.Error("Expected non-empty id")
	}
}

func TestCatalogService_CreateUnit_SaveError(t *testing.T) {
	repo := &mockUnitRepo{saveErr: errors.New("save error")}
	svc := NewCatalogService(nil, nil, repo, nil, nil)

	_, err := svc.CreateUnit(context.Background(), "prop-1", "ut-1", "Room 101")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestCatalogService_GetUnit_Success(t *testing.T) {
	unit := domain.NewUnit("prop-1", "ut-1", "Room 101")
	repo := &mockUnitRepo{unit: unit}
	svc := NewCatalogService(nil, nil, repo, nil, nil)

	result, err := svc.GetUnit(context.Background(), unit.ID())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Name() != "Room 101" {
		t.Errorf("Expected 'Room 101', got '%s'", result.Name())
	}
}

func TestCatalogService_GetUnit_NotFound(t *testing.T) {
	repo := &mockUnitRepo{unit: nil}
	svc := NewCatalogService(nil, nil, repo, nil, nil)

	_, err := svc.GetUnit(context.Background(), "non-existent")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCatalogService_ListUnits_Success(t *testing.T) {
	unit := domain.NewUnit("prop-1", "ut-1", "Room 101")
	repo := &mockUnitRepo{units: []*domain.Unit{unit}}
	svc := NewCatalogService(nil, nil, repo, nil, nil)

	units, count, err := svc.ListUnits(context.Background(), "prop-1", 1, 10)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(units) != 1 {
		t.Errorf("Expected 1 unit, got %d", len(units))
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestCatalogService_ListUnits_DefaultPagination(t *testing.T) {
	repo := &mockUnitRepo{units: []*domain.Unit{}}
	svc := NewCatalogService(nil, nil, repo, nil, nil)

	_, _, err := svc.ListUnits(context.Background(), "prop-1", 0, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_UpdateUnit_Success(t *testing.T) {
	unit := domain.NewUnit("prop-1", "ut-1", "Room 101")
	repo := &mockUnitRepo{unit: unit}
	svc := NewCatalogService(nil, nil, repo, nil, nil)

	err := svc.UpdateUnit(context.Background(), unit.ID(), "Room 202", "maintenance")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_UpdateUnit_NotFound(t *testing.T) {
	repo := &mockUnitRepo{unit: nil}
	svc := NewCatalogService(nil, nil, repo, nil, nil)

	err := svc.UpdateUnit(context.Background(), "non-existent", "Room 202", "maintenance")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCatalogService_DeleteUnit_Success(t *testing.T) {
	repo := &mockUnitRepo{}
	svc := NewCatalogService(nil, nil, repo, nil, nil)

	err := svc.DeleteUnit(context.Background(), "unit-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_CreateAmenity_Success(t *testing.T) {
	repo := &mockAmenityRepo{}
	svc := NewCatalogService(nil, nil, nil, repo, nil)

	id, err := svc.CreateAmenity(context.Background(), "WiFi", "High speed internet", "wifi-icon")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id == "" {
		t.Error("Expected non-empty id")
	}
}

func TestCatalogService_CreateAmenity_SaveError(t *testing.T) {
	repo := &mockAmenityRepo{saveErr: errors.New("save error")}
	svc := NewCatalogService(nil, nil, nil, repo, nil)

	_, err := svc.CreateAmenity(context.Background(), "WiFi", "desc", "icon")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestCatalogService_GetAmenity_Success(t *testing.T) {
	amenity, _ := domain.NewAmenity("WiFi", "desc", "icon")
	repo := &mockAmenityRepo{amenity: amenity}
	svc := NewCatalogService(nil, nil, nil, repo, nil)

	result, err := svc.GetAmenity(context.Background(), amenity.ID())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Name() != "WiFi" {
		t.Errorf("Expected 'WiFi', got '%s'", result.Name())
	}
}

func TestCatalogService_GetAmenity_NotFound(t *testing.T) {
	repo := &mockAmenityRepo{amenity: nil}
	svc := NewCatalogService(nil, nil, nil, repo, nil)

	_, err := svc.GetAmenity(context.Background(), "non-existent")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCatalogService_ListAmenities_Success(t *testing.T) {
	amenity, _ := domain.NewAmenity("WiFi", "desc", "icon")
	repo := &mockAmenityRepo{amenities: []*domain.Amenity{amenity}}
	svc := NewCatalogService(nil, nil, nil, repo, nil)

	amenities, count, err := svc.ListAmenities(context.Background(), 1, 10)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(amenities) != 1 {
		t.Errorf("Expected 1 amenity, got %d", len(amenities))
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestCatalogService_ListAmenities_DefaultPagination(t *testing.T) {
	repo := &mockAmenityRepo{amenities: []*domain.Amenity{}}
	svc := NewCatalogService(nil, nil, nil, repo, nil)

	_, _, err := svc.ListAmenities(context.Background(), 0, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_UpdateAmenity_Success(t *testing.T) {
	amenity, _ := domain.NewAmenity("WiFi", "desc", "icon")
	repo := &mockAmenityRepo{amenity: amenity}
	svc := NewCatalogService(nil, nil, nil, repo, nil)

	err := svc.UpdateAmenity(context.Background(), amenity.ID(), "High-Speed WiFi", "updated desc", "new-icon")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_UpdateAmenity_NotFound(t *testing.T) {
	repo := &mockAmenityRepo{amenity: nil}
	svc := NewCatalogService(nil, nil, nil, repo, nil)

	err := svc.UpdateAmenity(context.Background(), "non-existent", "name", "desc", "icon")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCatalogService_DeleteAmenity_Success(t *testing.T) {
	repo := &mockAmenityRepo{}
	svc := NewCatalogService(nil, nil, nil, repo, nil)

	err := svc.DeleteAmenity(context.Background(), "amenity-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_CreateGuestService_Success(t *testing.T) {
	repo := &mockGuestServiceRepo{}
	svc := NewCatalogService(nil, nil, nil, nil, repo)

	id, err := svc.CreateGuestService(context.Background(), "Spa", "Wellness center", "spa-icon")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id == "" {
		t.Error("Expected non-empty id")
	}
}

func TestCatalogService_CreateGuestService_SaveError(t *testing.T) {
	repo := &mockGuestServiceRepo{saveErr: errors.New("save error")}
	svc := NewCatalogService(nil, nil, nil, nil, repo)

	_, err := svc.CreateGuestService(context.Background(), "Spa", "desc", "icon")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestCatalogService_GetGuestService_Success(t *testing.T) {
	gs, _ := domain.NewGuestService("Spa", "desc", "icon")
	repo := &mockGuestServiceRepo{guestService: gs}
	svc := NewCatalogService(nil, nil, nil, nil, repo)

	result, err := svc.GetGuestService(context.Background(), gs.ID())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Name() != "Spa" {
		t.Errorf("Expected 'Spa', got '%s'", result.Name())
	}
}

func TestCatalogService_GetGuestService_NotFound(t *testing.T) {
	repo := &mockGuestServiceRepo{guestService: nil}
	svc := NewCatalogService(nil, nil, nil, nil, repo)

	_, err := svc.GetGuestService(context.Background(), "non-existent")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCatalogService_ListGuestServices_Success(t *testing.T) {
	gs, _ := domain.NewGuestService("Spa", "desc", "icon")
	repo := &mockGuestServiceRepo{guestServices: []*domain.GuestService{gs}}
	svc := NewCatalogService(nil, nil, nil, nil, repo)

	services, count, err := svc.ListGuestServices(context.Background(), 1, 10)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestCatalogService_ListGuestServices_DefaultPagination(t *testing.T) {
	repo := &mockGuestServiceRepo{guestServices: []*domain.GuestService{}}
	svc := NewCatalogService(nil, nil, nil, nil, repo)

	_, _, err := svc.ListGuestServices(context.Background(), 0, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_UpdateGuestService_Success(t *testing.T) {
	gs, _ := domain.NewGuestService("Spa", "desc", "icon")
	repo := &mockGuestServiceRepo{guestService: gs}
	svc := NewCatalogService(nil, nil, nil, nil, repo)

	err := svc.UpdateGuestService(context.Background(), gs.ID(), "Wellness Spa", "updated desc", "new-icon")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_UpdateGuestService_NotFound(t *testing.T) {
	repo := &mockGuestServiceRepo{guestService: nil}
	svc := NewCatalogService(nil, nil, nil, nil, repo)

	err := svc.UpdateGuestService(context.Background(), "non-existent", "name", "desc", "icon")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCatalogService_DeleteGuestService_Success(t *testing.T) {
	repo := &mockGuestServiceRepo{}
	svc := NewCatalogService(nil, nil, nil, nil, repo)

	err := svc.DeleteGuestService(context.Background(), "service-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_GetUnitType_RepoError(t *testing.T) {
	repo := &mockUnitTypeRepo{findErr: errors.New("repo error")}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	_, _, _, err := svc.GetUnitType(context.Background(), "ut-1")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_GetUnitTypeEntity_RepoError(t *testing.T) {
	repo := &mockUnitTypeRepo{findErr: errors.New("repo error")}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	_, err := svc.GetUnitTypeEntity(context.Background(), "ut-1")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_GetUnit_RepoError(t *testing.T) {
	repo := &mockUnitRepo{findErr: errors.New("repo error")}
	svc := NewCatalogService(nil, nil, repo, nil, nil)

	_, err := svc.GetUnit(context.Background(), "unit-1")
	if err == nil {
		t.Error("Expected error from repo")
	}}

func TestCatalogService_CreateProperty_RepoError(t *testing.T) {
	repo := &mockPropertyRepo{saveErr: errors.New("save error")}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	_, err := svc.CreateProperty(context.Background(), "org-1", "Hotel", "HTL", "hotel")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_UpdateUnitType_RepoError(t *testing.T) {
	ut, _ := domain.NewUnitType("prop-1", "Standard", "STD", 10, vo.NewMoney(15000, "USD"), 2, 2, 0, nil)
	repo := &mockUnitTypeRepo{unitType: ut}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	err := svc.UpdateUnitType(context.Background(), ut.ID(), "Deluxe", "DLX", 5, 19999, "USD", 2, 2, 1, []string{"wifi"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_UpdateUnitType_GetError(t *testing.T) {
	repo := &mockUnitTypeRepo{findErr: errors.New("find error")}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	err := svc.UpdateUnitType(context.Background(), "non-existent", "Deluxe", "DLX", 5, 19999, "USD", 2, 2, 1, nil)
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_UpdateUnitType_SaveError(t *testing.T) {
	ut, _ := domain.NewUnitType("prop-1", "Standard", "STD", 10, vo.NewMoney(15000, "USD"), 2, 2, 0, nil)
	repo := &mockUnitTypeRepo{unitType: ut, saveErr: errors.New("save error")}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	err := svc.UpdateUnitType(context.Background(), ut.ID(), "Deluxe", "DLX", 5, 19999, "USD", 2, 2, 1, nil)
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_UpdateUnit_RepoError(t *testing.T) {
	repo := &mockUnitRepo{findErr: errors.New("find error")}
	svc := NewCatalogService(nil, nil, repo, nil, nil)

	err := svc.UpdateUnit(context.Background(), "non-existent", "New Unit", "available")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_UpdateUnit_SaveError(t *testing.T) {
	u := domain.NewUnit("prop-1", "ut-1", "Unit 101")
	repo := &mockUnitRepo{unit: u, saveErr: errors.New("save error")}
	svc := NewCatalogService(nil, nil, repo, nil, nil)

	err := svc.UpdateUnit(context.Background(), u.ID(), "New Name", "available")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_UpdateProperty_RepoError(t *testing.T) {
	repo := &mockPropertyRepo{findErr: errors.New("find error")}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	err := svc.UpdateProperty(context.Background(), "non-existent", "Name", "CODE", "hotel")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_UpdateProperty_SaveError(t *testing.T) {
	prop := domain.ReconstituteProperty("prop-1", "org-1", "Old Name", "OLD", "hotel", time.Now())
	repo := &mockPropertyRepo{prop: prop, saveErr: errors.New("save error")}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	err := svc.UpdateProperty(context.Background(), "prop-1", "New Name", "NEW", "hotel")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_UpdateAmenity_RepoError(t *testing.T) {
	a := domain.ReconstituteAmenity("a-1", "WiFi", "Wireless", "wifi-icon")
	repo := &mockAmenityRepo{amenity: a, saveErr: errors.New("save error")}
	svc := NewCatalogService(nil, nil, nil, repo, nil)

	err := svc.UpdateAmenity(context.Background(), "a-1", "New WiFi", "New desc", "new-icon")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_UpdateAmenity_GetError(t *testing.T) {
	repo := &mockAmenityRepo{findErr: errors.New("find error")}
	svc := NewCatalogService(nil, nil, nil, repo, nil)

	err := svc.UpdateAmenity(context.Background(), "non-existent", "WiFi", "desc", "icon")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_UpdateGuestService_RepoError(t *testing.T) {
	gs := domain.ReconstituteGuestService("gs-1", "Spa", "Spa service", "spa-icon")
	repo := &mockGuestServiceRepo{guestService: gs, saveErr: errors.New("save error")}
	svc := NewCatalogService(nil, nil, nil, nil, repo)

	err := svc.UpdateGuestService(context.Background(), "gs-1", "New Spa", "New desc", "new-icon")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_UpdateGuestService_GetError(t *testing.T) {
	repo := &mockGuestServiceRepo{findErr: errors.New("find error")}
	svc := NewCatalogService(nil, nil, nil, nil, repo)

	err := svc.UpdateGuestService(context.Background(), "non-existent", "Spa", "desc", "icon")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_DeleteGuestService_RepoError(t *testing.T) {
	repo := &mockGuestServiceRepo{deleteErr: errors.New("delete error")}
	svc := NewCatalogService(nil, nil, nil, nil, repo)

	err := svc.DeleteGuestService(context.Background(), "gs-1")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_CreateProperty_DefaultType(t *testing.T) {
	repo := &mockPropertyRepo{}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	id, err := svc.CreateProperty(context.Background(), "org-1", "Hotel", "HTL", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id == "" {
		t.Error("Expected non-empty id")
	}
}

func TestCatalogService_CreateProperty_NewPropertyError(t *testing.T) {
	repo := &mockPropertyRepo{}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	_, err := svc.CreateProperty(context.Background(), "", "", "", "invalid")
	if err == nil {
		t.Error("Expected error from NewProperty")
	}
}

func TestCatalogService_CreateUnitType_NewUnitTypeError(t *testing.T) {
	repo := &mockUnitTypeRepo{}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	_, err := svc.CreateUnitType(context.Background(), "", "", "", 0, 0, "", 0, 0, 0, nil)
	if err == nil {
		t.Error("Expected error from NewUnitType")
	}
}

func TestCatalogService_CreateAmenity_NewAmenityError(t *testing.T) {
	repo := &mockAmenityRepo{}
	svc := NewCatalogService(nil, nil, nil, repo, nil)

	_, err := svc.CreateAmenity(context.Background(), "", "", "")
	if err == nil {
		t.Error("Expected error from NewAmenity")
	}
}

func TestCatalogService_CreateGuestService_NewGuestServiceError(t *testing.T) {
	repo := &mockGuestServiceRepo{}
	svc := NewCatalogService(nil, nil, nil, nil, repo)

	_, err := svc.CreateGuestService(context.Background(), "", "", "")
	if err == nil {
		t.Error("Expected error from NewGuestService")
	}
}
