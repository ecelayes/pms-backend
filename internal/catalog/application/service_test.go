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
	saveErr          error
	findByIDResult   *domain.Property
	findByIDErr      error
	findByOrgResult  []*domain.Property
	findByOrgTotal   int64
	findByOrgErr     error
	deleteErr        error
}

func (m *mockPropertyRepo) SaveProperty(ctx context.Context, p *domain.Property) error {
	return m.saveErr
}
func (m *mockPropertyRepo) FindPropertyByID(ctx context.Context, id string) (*domain.Property, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	return m.findByIDResult, nil
}
func (m *mockPropertyRepo) FindPropertiesByOrganization(ctx context.Context, orgID string, limit int, offset int) ([]*domain.Property, int64, error) {
	if m.findByOrgErr != nil {
		return nil, 0, m.findByOrgErr
	}
	return m.findByOrgResult, m.findByOrgTotal, nil
}
func (m *mockPropertyRepo) DeleteProperty(ctx context.Context, id string) error {
	return m.deleteErr
}

type mockUnitTypeRepo struct {
	findByIDResult *domain.UnitType
	findByIDErr    error
	saveErr        error
	deleteErr      error
}

func (m *mockUnitTypeRepo) SaveUnitType(ctx context.Context, ut *domain.UnitType) error {
	return m.saveErr
}
func (m *mockUnitTypeRepo) FindUnitTypeByID(ctx context.Context, id string) (*domain.UnitType, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	return m.findByIDResult, nil
}
func (m *mockUnitTypeRepo) FindUnitTypesByPropertyID(ctx context.Context, propertyID string, limit, offset int) ([]*domain.UnitType, int64, error) {
	return nil, 0, nil
}
func (m *mockUnitTypeRepo) DeleteUnitType(ctx context.Context, id string) error {
	return m.deleteErr
}

type mockUnitRepo struct {
	saveErr   error
	findErr   error
	deleteErr error
}

func (m *mockUnitRepo) SaveUnit(ctx context.Context, u *domain.Unit) error {
	return m.saveErr
}
func (m *mockUnitRepo) FindUnitByID(ctx context.Context, id string) (*domain.Unit, error) {
	return nil, m.findErr
}
func (m *mockUnitRepo) FindUnitsByPropertyID(ctx context.Context, propertyID string, limit, offset int) ([]*domain.Unit, int64, error) {
	return nil, 0, nil
}
func (m *mockUnitRepo) DeleteUnit(ctx context.Context, id string) error {
	return m.deleteErr
}

type mockAmenityRepo struct {
	saveErr error
	listErr error
}

func (m *mockAmenityRepo) SaveAmenity(ctx context.Context, a *domain.Amenity) error {
	return m.saveErr
}
func (m *mockAmenityRepo) ListAmenities(ctx context.Context) ([]*domain.Amenity, error) {
	return nil, m.listErr
}

type mockGuestServiceRepo struct {
	saveErr error
	listErr error
}

func (m *mockGuestServiceRepo) SaveGuestService(ctx context.Context, gs *domain.GuestService) error {
	return m.saveErr
}
func (m *mockGuestServiceRepo) ListGuestServices(ctx context.Context) ([]*domain.GuestService, error) {
	return nil, m.listErr
}


func TestNewCatalogService(t *testing.T) {
	svc := NewCatalogService(nil, nil, nil, nil, nil)
	if svc == nil {
		t.Error("Expected non-nil service")
	}
}

func TestCatalogService_CreateProperty(t *testing.T) {
	repo := &mockPropertyRepo{}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	id, err := svc.CreateProperty(context.Background(), "org-1", "Test Hotel", "TH-001", "hotel")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id == "" {
		t.Error("Expected non-empty ID")
	}
}

func TestCatalogService_CreateProperty_EmptyName(t *testing.T) {
	repo := &mockPropertyRepo{}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	_, err := svc.CreateProperty(context.Background(), "org-1", "", "TH-001", "hotel")
	if err == nil {
		t.Error("Expected error for empty name")
	}
}

func TestCatalogService_CreateProperty_EmptyCode(t *testing.T) {
	repo := &mockPropertyRepo{}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	_, err := svc.CreateProperty(context.Background(), "org-1", "Test Hotel", "", "hotel")
	if err == nil {
		t.Error("Expected error for empty code")
	}
}

func TestCatalogService_CreateProperty_RepoError(t *testing.T) {
	repo := &mockPropertyRepo{saveErr: errors.New("save error")}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	_, err := svc.CreateProperty(context.Background(), "org-1", "Test Hotel", "TH-001", "hotel")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_GetProperty(t *testing.T) {
	prop := domain.ReconstituteProperty("prop-1", "org-1", "Test Hotel", "TH-001", "hotel", time.Now())
	repo := &mockPropertyRepo{findByIDResult: prop}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	result, err := svc.GetProperty(context.Background(), "prop-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.ID() != "prop-1" {
		t.Errorf("Expected prop-1, got %s", result.ID())
	}
}

func TestCatalogService_GetProperty_NotFound(t *testing.T) {
	repo := &mockPropertyRepo{findByIDResult: nil}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	_, err := svc.GetProperty(context.Background(), "non-existent")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
}

func TestCatalogService_GetProperty_RepoError(t *testing.T) {
	repo := &mockPropertyRepo{findByIDErr: errors.New("repo error")}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	_, err := svc.GetProperty(context.Background(), "prop-1")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_ListProperties(t *testing.T) {
	prop1 := domain.ReconstituteProperty("prop-1", "org-1", "Hotel 1", "H1", "hotel", time.Now())
	prop2 := domain.ReconstituteProperty("prop-2", "org-1", "Hotel 2", "H2", "hotel", time.Now())
	repo := &mockPropertyRepo{findByOrgResult: []*domain.Property{prop1, prop2}, findByOrgTotal: 2}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	results, total, err := svc.ListProperties(context.Background(), "org-1", 1, 10)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if total != 2 {
		t.Errorf("Expected total 2, got %d", total)
	}
}

func TestCatalogService_ListProperties_DefaultPagination(t *testing.T) {
	repo := &mockPropertyRepo{findByOrgResult: []*domain.Property{}}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	// page and limit < 1 should be defaulted to 1 and 10
	_, _, err := svc.ListProperties(context.Background(), "org-1", 0, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_UpdateProperty(t *testing.T) {
	prop := domain.ReconstituteProperty("prop-1", "org-1", "Original", "ORIG", "hotel", time.Now())
	repo := &mockPropertyRepo{findByIDResult: prop}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	err := svc.UpdateProperty(context.Background(), "prop-1", "Updated", "UPD", "apartment")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if prop.Name() != "Updated" {
		t.Errorf("Expected name 'Updated', got '%s'", prop.Name())
	}
	if prop.Code() != "UPD" {
		t.Errorf("Expected code 'UPD', got '%s'", prop.Code())
	}
}

func TestCatalogService_UpdateProperty_NotFound(t *testing.T) {
	repo := &mockPropertyRepo{findByIDResult: nil}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	err := svc.UpdateProperty(context.Background(), "non-existent", "Updated", "UPD", "hotel")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
}

func TestCatalogService_DeleteProperty(t *testing.T) {
	repo := &mockPropertyRepo{}
	svc := NewCatalogService(nil, repo, nil, nil, nil)

	err := svc.DeleteProperty(context.Background(), "prop-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCatalogService_GetUnitType(t *testing.T) {
	ut := domain.ReconstituteUnitType(
		"ut-1", "prop-1", "Standard Room", "SR-001",
		10, vo.NewMoney(15000, "USD"),
		2, 2, 1, []string{}, time.Now(),
	)
	repo := &mockUnitTypeRepo{findByIDResult: ut}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	propID, basePrice, totalQty, err := svc.GetUnitType(context.Background(), "ut-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if propID != "prop-1" {
		t.Errorf("Expected prop-1, got %s", propID)
	}
	if basePrice.Amount() != 15000 {
		t.Errorf("Expected 15000, got %d", basePrice.Amount())
	}
	if totalQty != 10 {
		t.Errorf("Expected 10, got %d", totalQty)
	}
}

func TestCatalogService_GetUnitType_NotFound(t *testing.T) {
	repo := &mockUnitTypeRepo{findByIDResult: nil}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	_, _, _, err := svc.GetUnitType(context.Background(), "non-existent")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
}

func TestCatalogService_GetUnitType_RepoError(t *testing.T) {
	repo := &mockUnitTypeRepo{findByIDErr: errors.New("repo error")}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	_, _, _, err := svc.GetUnitType(context.Background(), "ut-1")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestCatalogService_GetUnitTypeEntity(t *testing.T) {
	ut := domain.ReconstituteUnitType(
		"ut-1", "prop-1", "Standard Room", "SR-001",
		10, vo.NewMoney(15000, "USD"),
		2, 2, 1, []string{}, time.Now(),
	)
	repo := &mockUnitTypeRepo{findByIDResult: ut}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	result, err := svc.GetUnitTypeEntity(context.Background(), "ut-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.ID() != "ut-1" {
		t.Errorf("Expected ut-1, got %s", result.ID())
	}
}

func TestCatalogService_GetUnitTypeEntity_NotFound(t *testing.T) {
	repo := &mockUnitTypeRepo{findByIDResult: nil}
	svc := NewCatalogService(repo, nil, nil, nil, nil)

	_, err := svc.GetUnitTypeEntity(context.Background(), "non-existent")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
}
