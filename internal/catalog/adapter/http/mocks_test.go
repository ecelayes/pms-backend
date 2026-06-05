package http

import (
	"context"

	"github.com/ecelayes/pms-backend/internal/catalog/domain"
)

// --- PropertyService mock ---

type mockPropertyService struct {
	createResult  string
	createErr     error
	getResult     *domain.Property
	getErr        error
	listResult    []*domain.Property
	listTotal     int64
	listErr       error
	updateErr     error
	deleteErr     error
}

func (m *mockPropertyService) CreateProperty(ctx context.Context, orgID, name, code, pType string) (string, error) {
	return m.createResult, m.createErr
}
func (m *mockPropertyService) GetProperty(ctx context.Context, id string) (*domain.Property, error) {
	return m.getResult, m.getErr
}
func (m *mockPropertyService) ListProperties(ctx context.Context, orgID string, page, limit int) ([]*domain.Property, int64, error) {
	return m.listResult, m.listTotal, m.listErr
}
func (m *mockPropertyService) UpdateProperty(ctx context.Context, id, name, code, pType string) error {
	return m.updateErr
}
func (m *mockPropertyService) DeleteProperty(ctx context.Context, id string) error {
	return m.deleteErr
}

// --- UnitTypeService mock ---

type mockUnitTypeService struct {
	createResult  string
	createErr     error
	getResult     *domain.UnitType
	getErr        error
	listResult    []*domain.UnitType
	listTotal     int64
	listErr       error
	updateErr     error
	deleteErr     error
}

func (m *mockUnitTypeService) CreateUnitType(ctx context.Context, propertyID, name, code string, qty int, priceCents int64, currency string, maxOcc, maxAd, maxCh int, amenities []string) (string, error) {
	return m.createResult, m.createErr
}
func (m *mockUnitTypeService) ListUnitTypes(ctx context.Context, propertyID string, page, limit int) ([]*domain.UnitType, int64, error) {
	return m.listResult, m.listTotal, m.listErr
}
func (m *mockUnitTypeService) GetUnitTypeEntity(ctx context.Context, id string) (*domain.UnitType, error) {
	return m.getResult, m.getErr
}
func (m *mockUnitTypeService) UpdateUnitType(ctx context.Context, id, name, code string, qty int, priceCents int64, currency string, maxOcc, maxAd, maxCh int, amenities []string) error {
	return m.updateErr
}
func (m *mockUnitTypeService) DeleteUnitType(ctx context.Context, id string) error {
	return m.deleteErr
}

// --- UnitService mock ---

type mockUnitService struct {
	createResult string
	createErr    error
	getResult    *domain.Unit
	getErr       error
	listResult   []*domain.Unit
	listTotal    int64
	listErr      error
	updateErr    error
	deleteErr    error
}

func (m *mockUnitService) CreateUnit(ctx context.Context, propertyID, unitTypeID, name string) (string, error) {
	return m.createResult, m.createErr
}
func (m *mockUnitService) GetUnit(ctx context.Context, id string) (*domain.Unit, error) {
	return m.getResult, m.getErr
}
func (m *mockUnitService) ListUnits(ctx context.Context, propertyID string, page, limit int) ([]*domain.Unit, int64, error) {
	return m.listResult, m.listTotal, m.listErr
}
func (m *mockUnitService) UpdateUnit(ctx context.Context, id, name, status string) error {
	return m.updateErr
}
func (m *mockUnitService) DeleteUnit(ctx context.Context, id string) error {
	return m.deleteErr
}

// --- CatalogService (amenities + services) mock ---

type mockCatalogService struct {
	amenityCreateResult    string
	amenityCreateErr       error
	amenityGetResult       *domain.Amenity
	amenityGetErr          error
	amenityListResult      []*domain.Amenity
	amenityListTotal       int64
	amenityListErr         error
	amenityUpdateErr       error
	amenityDeleteErr       error
	guestCreateResult      string
	guestCreateErr         error
	guestGetResult         *domain.GuestService
	guestGetErr            error
	guestListResult        []*domain.GuestService
	guestListTotal         int64
	guestListErr           error
	guestUpdateErr         error
	guestDeleteErr         error
}

func (m *mockCatalogService) CreateAmenity(ctx context.Context, name, description, icon string) (string, error) {
	return m.amenityCreateResult, m.amenityCreateErr
}
func (m *mockCatalogService) GetAmenity(ctx context.Context, id string) (*domain.Amenity, error) {
	return m.amenityGetResult, m.amenityGetErr
}
func (m *mockCatalogService) ListAmenities(ctx context.Context, page, limit int) ([]*domain.Amenity, int64, error) {
	return m.amenityListResult, m.amenityListTotal, m.amenityListErr
}
func (m *mockCatalogService) UpdateAmenity(ctx context.Context, id, name, description, icon string) error {
	return m.amenityUpdateErr
}
func (m *mockCatalogService) DeleteAmenity(ctx context.Context, id string) error {
	return m.amenityDeleteErr
}
func (m *mockCatalogService) CreateGuestService(ctx context.Context, name, description, icon string) (string, error) {
	return m.guestCreateResult, m.guestCreateErr
}
func (m *mockCatalogService) GetGuestService(ctx context.Context, id string) (*domain.GuestService, error) {
	return m.guestGetResult, m.guestGetErr
}
func (m *mockCatalogService) ListGuestServices(ctx context.Context, page, limit int) ([]*domain.GuestService, int64, error) {
	return m.guestListResult, m.guestListTotal, m.guestListErr
}
func (m *mockCatalogService) UpdateGuestService(ctx context.Context, id, name, description, icon string) error {
	return m.guestUpdateErr
}
func (m *mockCatalogService) DeleteGuestService(ctx context.Context, id string) error {
	return m.guestDeleteErr
}
