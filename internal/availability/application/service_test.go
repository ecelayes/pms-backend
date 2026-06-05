package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ecelayes/pms-backend/internal/catalog/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
)

type mockAvailabilityRepo struct {
	inventory          map[string]int
	inventoryErr      error
	batchInventory    map[string]int
	batchInventoryErr error
	updateInventoryErr error
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

type mockCatalogPort struct {
	unitTypes []*domain.UnitType
	err      error
}

func (m *mockCatalogPort) ListUnitTypes(ctx context.Context, propertyID string, page, limit int) ([]*domain.UnitType, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	return m.unitTypes, int64(len(m.unitTypes)), nil
}

type mockPricingPort struct {
	price  vo.Money
	err    error
}

func (m *mockPricingPort) CalculateBasePrice(ctx context.Context, unitTypeID string, defaultPrice vo.Money, start, end time.Time) (vo.Money, error) {
	if m.err != nil {
		return vo.Money{}, m.err
	}
	return m.price, nil
}

func futureDate(daysAhead int) time.Time {
	return time.Now().AddDate(0, 0, daysAhead)
}

// UpdateInventory Tests
func TestAvailabilityService_UpdateInventory_Success(t *testing.T) {
	repo := &mockAvailabilityRepo{}
	svc := NewAvailabilityService(repo, nil, nil)

	start := futureDate(1)
	end := futureDate(3)

	err := svc.UpdateInventory(context.Background(), "prop-1", "ut-1", start, end, -1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestAvailabilityService_UpdateInventory_RepoError(t *testing.T) {
	repo := &mockAvailabilityRepo{updateInventoryErr: errors.New("update error")}
	svc := NewAvailabilityService(repo, nil, nil)

	start := futureDate(1)
	end := futureDate(3)

	err := svc.UpdateInventory(context.Background(), "prop-1", "ut-1", start, end, -1)
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestAvailabilityService_UpdateInventory_SingleDay(t *testing.T) {
	repo := &mockAvailabilityRepo{}
	svc := NewAvailabilityService(repo, nil, nil)

	start := futureDate(1)
	end := futureDate(2)

	err := svc.UpdateInventory(context.Background(), "prop-1", "ut-1", start, end, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestAvailabilityService_UpdateInventory_ZeroDelta(t *testing.T) {
	repo := &mockAvailabilityRepo{}
	svc := NewAvailabilityService(repo, nil, nil)

	start := futureDate(1)
	end := futureDate(3)

	err := svc.UpdateInventory(context.Background(), "prop-1", "ut-1", start, end, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// Search Tests
func TestAvailabilityService_Search_Success(t *testing.T) {
	unitType := domain.ReconstituteUnitType("ut-1", "prop-1", "Standard", "STD", 10, vo.NewMoney(10000, "USD"), 2, 2, 1, nil, time.Now())
	
	repo := &mockAvailabilityRepo{
		batchInventory: map[string]int{
			futureDate(1).Format("2006-01-02"): 5,
			futureDate(2).Format("2006-01-02"): 5,
		},
	}
	catalog := &mockCatalogPort{unitTypes: []*domain.UnitType{unitType}}
	pricing := &mockPricingPort{price: vo.NewMoney(20000, "USD")}
	
	svc := NewAvailabilityService(repo, catalog, pricing)

	start := futureDate(1)
	end := futureDate(3)

	results, err := svc.Search(context.Background(), "prop-1", start, end, 2, 0, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if results[0].UnitTypeID != "ut-1" {
		t.Errorf("Expected ut-1, got %s", results[0].UnitTypeID)
	}
}

func TestAvailabilityService_Search_CatalogError(t *testing.T) {
	repo := &mockAvailabilityRepo{}
	catalog := &mockCatalogPort{err: errors.New("catalog error")}
	pricing := &mockPricingPort{}
	
	svc := NewAvailabilityService(repo, catalog, pricing)

	start := futureDate(1)
	end := futureDate(3)

	results, err := svc.Search(context.Background(), "prop-1", start, end, 2, 0, 1)
	if err == nil {
		t.Error("Expected error from catalog")
	}
	if results != nil {
		t.Error("Expected nil results on error")
	}
}

func TestAvailabilityService_Search_NoUnitTypes(t *testing.T) {
	repo := &mockAvailabilityRepo{}
	catalog := &mockCatalogPort{unitTypes: nil}
	pricing := &mockPricingPort{}
	
	svc := NewAvailabilityService(repo, catalog, pricing)

	start := futureDate(1)
	end := futureDate(3)

	results, err := svc.Search(context.Background(), "prop-1", start, end, 2, 0, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestAvailabilityService_Search_EmptyDates(t *testing.T) {
	repo := &mockAvailabilityRepo{}
	catalog := &mockCatalogPort{unitTypes: []*domain.UnitType{}}
	pricing := &mockPricingPort{}
	
	svc := NewAvailabilityService(repo, catalog, pricing)

	start := futureDate(1)
	end := futureDate(1) // Same day = no nights

	results, err := svc.Search(context.Background(), "prop-1", start, end, 2, 0, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty dates, got %d", len(results))
	}
}

func TestAvailabilityService_Search_UnitTypeExcludedByAdults(t *testing.T) {
	unitType := domain.ReconstituteUnitType("ut-1", "prop-1", "Standard", "STD", 10, vo.NewMoney(10000, "USD"), 2, 2, 0, nil, time.Now()) // max 2 adults
	
	repo := &mockAvailabilityRepo{}
	catalog := &mockCatalogPort{unitTypes: []*domain.UnitType{unitType}}
	pricing := &mockPricingPort{}
	
	svc := NewAvailabilityService(repo, catalog, pricing)

	start := futureDate(1)
	end := futureDate(3)

	// Request 3 adults - should be excluded
	results, err := svc.Search(context.Background(), "prop-1", start, end, 3, 0, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results (unit max 2 adults), got %d", len(results))
	}
}

func TestAvailabilityService_Search_UnitTypeExcludedByTotalOccupancy(t *testing.T) {
	unitType := domain.ReconstituteUnitType("ut-1", "prop-1", "Standard", "STD", 10, vo.NewMoney(10000, "USD"), 2, 2, 0, nil, time.Now()) // max 2 adults + 2 children = 4 total
	
	repo := &mockAvailabilityRepo{}
	catalog := &mockCatalogPort{unitTypes: []*domain.UnitType{unitType}}
	pricing := &mockPricingPort{}
	
	svc := NewAvailabilityService(repo, catalog, pricing)

	start := futureDate(1)
	end := futureDate(3)

	// Request 2 adults + 3 children = 5 total - should be excluded
	results, err := svc.Search(context.Background(), "prop-1", start, end, 2, 3, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results (max 4 occupancy), got %d", len(results))
	}
}

func TestAvailabilityService_Search_NoAvailability(t *testing.T) {
	unitType := domain.ReconstituteUnitType("ut-1", "prop-1", "Standard", "STD", 10, vo.NewMoney(10000, "USD"), 2, 2, 1, nil, time.Now())
	
	repo := &mockAvailabilityRepo{
		batchInventory: map[string]int{
			futureDate(1).Format("2006-01-02"): 0,
			futureDate(2).Format("2006-01-02"): 0,
		},
	}
	catalog := &mockCatalogPort{unitTypes: []*domain.UnitType{unitType}}
	pricing := &mockPricingPort{}
	
	svc := NewAvailabilityService(repo, catalog, pricing)

	start := futureDate(1)
	end := futureDate(3)

	results, err := svc.Search(context.Background(), "prop-1", start, end, 2, 0, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results (no availability), got %d", len(results))
	}
}

func TestAvailabilityService_Search_NotEnoughRooms(t *testing.T) {
	unitType := domain.ReconstituteUnitType("ut-1", "prop-1", "Standard", "STD", 10, vo.NewMoney(10000, "USD"), 2, 2, 1, nil, time.Now())
	
	repo := &mockAvailabilityRepo{
		batchInventory: map[string]int{
			futureDate(1).Format("2006-01-02"): 5,
			futureDate(2).Format("2006-01-02"): 5,
		},
	}
	catalog := &mockCatalogPort{unitTypes: []*domain.UnitType{unitType}}
	pricing := &mockPricingPort{}
	
	svc := NewAvailabilityService(repo, catalog, pricing)

	start := futureDate(1)
	end := futureDate(3)

	// Request 10 rooms but only 5 available
	results, err := svc.Search(context.Background(), "prop-1", start, end, 2, 0, 10)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results (only 5 rooms), got %d", len(results))
	}
}

func TestAvailabilityService_Search_PricingError(t *testing.T) {
	unitType := domain.ReconstituteUnitType("ut-1", "prop-1", "Standard", "STD", 10, vo.NewMoney(10000, "USD"), 2, 2, 1, nil, time.Now())
	
	repo := &mockAvailabilityRepo{
		batchInventory: map[string]int{
			futureDate(1).Format("2006-01-02"): 5,
			futureDate(2).Format("2006-01-02"): 5,
		},
	}
	catalog := &mockCatalogPort{unitTypes: []*domain.UnitType{unitType}}
	pricing := &mockPricingPort{err: errors.New("pricing error")}
	
	svc := NewAvailabilityService(repo, catalog, pricing)

	start := futureDate(1)
	end := futureDate(3)

	results, err := svc.Search(context.Background(), "prop-1", start, end, 2, 0, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Pricing error should skip the unit type, resulting in 0 results
	if len(results) != 0 {
		t.Errorf("Expected 0 results (pricing error), got %d", len(results))
	}
}

func TestAvailabilityService_Search_ZeroPrice(t *testing.T) {
	unitType := domain.ReconstituteUnitType("ut-1", "prop-1", "Standard", "STD", 10, vo.NewMoney(10000, "USD"), 2, 2, 1, nil, time.Now())
	
	repo := &mockAvailabilityRepo{
		batchInventory: map[string]int{
			futureDate(1).Format("2006-01-02"): 5,
			futureDate(2).Format("2006-01-02"): 5,
		},
	}
	catalog := &mockCatalogPort{unitTypes: []*domain.UnitType{unitType}}
	pricing := &mockPricingPort{price: vo.NewMoney(0, "USD")}
	
	svc := NewAvailabilityService(repo, catalog, pricing)

	start := futureDate(1)
	end := futureDate(3)

	results, err := svc.Search(context.Background(), "prop-1", start, end, 2, 0, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Zero price should skip the unit type
	if len(results) != 0 {
		t.Errorf("Expected 0 results (zero price), got %d", len(results))
	}
}

func TestAvailabilityService_Search_BatchInventoryError(t *testing.T) {
	unitType := domain.ReconstituteUnitType("ut-1", "prop-1", "Standard", "STD", 10, vo.NewMoney(10000, "USD"), 2, 2, 1, nil, time.Now())
	
	repo := &mockAvailabilityRepo{
		batchInventoryErr: errors.New("batch error"),
	}
	catalog := &mockCatalogPort{unitTypes: []*domain.UnitType{unitType}}
	pricing := &mockPricingPort{}
	
	svc := NewAvailabilityService(repo, catalog, pricing)

	start := futureDate(1)
	end := futureDate(3)

	results, err := svc.Search(context.Background(), "prop-1", start, end, 2, 0, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Batch error should skip the unit type
	if len(results) != 0 {
		t.Errorf("Expected 0 results (batch error), got %d", len(results))
	}
}

func TestAvailabilityService_Search_MultipleUnitTypes(t *testing.T) {
	unitType1 := domain.ReconstituteUnitType("ut-1", "prop-1", "Standard", "STD", 10, vo.NewMoney(10000, "USD"), 2, 2, 1, nil, time.Now())
	unitType2 := domain.ReconstituteUnitType("ut-2", "prop-1", "Deluxe", "DLX", 5, vo.NewMoney(20000, "USD"), 2, 2, 1, nil, time.Now())
	
	repo := &mockAvailabilityRepo{
		batchInventory: map[string]int{
			futureDate(1).Format("2006-01-02"): 5,
			futureDate(2).Format("2006-01-02"): 5,
		},
	}
	catalog := &mockCatalogPort{unitTypes: []*domain.UnitType{unitType1, unitType2}}
	pricing := &mockPricingPort{price: vo.NewMoney(20000, "USD")}
	
	svc := NewAvailabilityService(repo, catalog, pricing)

	start := futureDate(1)
	end := futureDate(3)

	results, err := svc.Search(context.Background(), "prop-1", start, end, 2, 0, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestAvailabilityService_Search_DefaultInventory(t *testing.T) {
	unitType := domain.ReconstituteUnitType("ut-1", "prop-1", "Standard", "STD", 10, vo.NewMoney(10000, "USD"), 2, 2, 1, nil, time.Now())
	
	// No inventory in repo - should use unit type's TotalQuantity
	repo := &mockAvailabilityRepo{
		batchInventory: map[string]int{}, // empty - should default to TotalQuantity
	}
	catalog := &mockCatalogPort{unitTypes: []*domain.UnitType{unitType}}
	pricing := &mockPricingPort{price: vo.NewMoney(20000, "USD")}
	
	svc := NewAvailabilityService(repo, catalog, pricing)

	start := futureDate(1)
	end := futureDate(3)

	results, err := svc.Search(context.Background(), "prop-1", start, end, 2, 0, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Should have 1 result since no inventory means using default TotalQuantity
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestAvailabilityService_Search_TrulyEmptyDates(t *testing.T) {
	repo := &mockAvailabilityRepo{}
	catalog := &mockCatalogPort{unitTypes: nil}
	pricing := &mockPricingPort{}
	
	svc := NewAvailabilityService(repo, catalog, pricing)

	// Same exact time = no dates
	t1 := futureDate(5)
	results, err := svc.Search(context.Background(), "prop-1", t1, t1, 2, 0, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if results != nil && len(results) != 0 {
		t.Errorf("Expected empty results")
	}
}
