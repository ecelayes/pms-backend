package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ecelayes/pms-backend/internal/shared/vo"
)


type mockAvailabilityRepo struct {
	inventory          map[string]int
	inventoryErr        error
	batchInventory      map[string]int
	batchInventoryErr   error
	updateInventoryErr   error
}

func (m *mockAvailabilityRepo) UpdateInventory(ctx context.Context, propertyID, unitID string, date time.Time, delta int) error {
	if m.updateInventoryErr != nil {
		return m.updateInventoryErr
	}
	return nil
}
func (m *mockAvailabilityRepo) GetInventory(ctx context.Context, propertyID, unitID string, date time.Time) (int, error) {
	if m.inventoryErr != nil {
		return 0, m.inventoryErr
	}
	key := propertyID + ":" + unitID + ":" + date.Format("2006-01-02")
	if val, ok := m.inventory[key]; ok {
		return val, nil
	}
	return 0, nil
}
func (m *mockAvailabilityRepo) GetBatchInventory(ctx context.Context, propertyID, unitID string, dates []time.Time) (map[string]int, error) {
	if m.batchInventoryErr != nil {
		return nil, m.batchInventoryErr
	}
	return m.batchInventory, nil
}

type mockCatalogUnitType struct {
	id            string
	propertyID    string
	name          string
	basePrice     vo.Money
	totalQuantity int
	maxAdults     int
	maxChildren   int
}

func (m *mockCatalogUnitType) ID() string         { return m.id }
func (m *mockCatalogUnitType) PropertyID() string  { return m.propertyID }
func (m *mockCatalogUnitType) Name() string       { return m.name }
func (m *mockCatalogUnitType) BasePrice() vo.Money { return m.basePrice }
func (m *mockCatalogUnitType) TotalQuantity() int { return m.totalQuantity }
func (m *mockCatalogUnitType) MaxAdults() int     { return m.maxAdults }
func (m *mockCatalogUnitType) MaxChildren() int   { return m.maxChildren }

type mockCatalogServiceInterface interface {
	ListUnitTypes(ctx context.Context, propertyID string, page, limit int) ([]*mockCatalogUnitType, int64, error)
}

type mockPricingServiceInterface interface {
	CalculateBasePrice(ctx context.Context, unitTypeID string, defaultPrice vo.Money, start, end time.Time) (vo.Money, error)
}


func TestAvailabilityService_UpdateInventory_Success(t *testing.T) {
	repo := &mockAvailabilityRepo{}
	svc := NewAvailabilityService(repo, nil, nil)

	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)

	err := svc.UpdateInventory(context.Background(), "prop-1", "ut-1", start, end, -1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestAvailabilityService_UpdateInventory_RepoError(t *testing.T) {
	repo := &mockAvailabilityRepo{updateInventoryErr: errors.New("update error")}
	svc := NewAvailabilityService(repo, nil, nil)

	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)

	err := svc.UpdateInventory(context.Background(), "prop-1", "ut-1", start, end, -1)
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestAvailabilityService_UpdateInventory_SingleDay(t *testing.T) {
	repo := &mockAvailabilityRepo{}
	svc := NewAvailabilityService(repo, nil, nil)

	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)

	err := svc.UpdateInventory(context.Background(), "prop-1", "ut-1", start, end, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestAvailabilityService_UpdateInventory_ZeroDelta(t *testing.T) {
	repo := &mockAvailabilityRepo{}
	svc := NewAvailabilityService(repo, nil, nil)

	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)

	err := svc.UpdateInventory(context.Background(), "prop-1", "ut-1", start, end, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
