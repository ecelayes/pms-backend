package domain

import (
	"testing"
	"time"

	"github.com/ecelayes/pms-backend/internal/shared/vo"
)

func TestNewPriceRule(t *testing.T) {
	startDate, _ := time.Parse("2006-01-02", "2024-06-01")
	endDate, _ := time.Parse("2006-01-02", "2024-06-30")
	dateRange, _ := vo.NewDateRange(startDate, endDate)
	price := vo.NewMoney(15000, "USD")
	
	rule := NewPriceRule("unit-type-123", dateRange, price)
	
	if rule.ID() == "" {
		t.Error("Expected non-empty ID")
	}
	if rule.UnitTypeID() != "unit-type-123" {
		t.Errorf("Expected UnitTypeID 'unit-type-123', got '%s'", rule.UnitTypeID())
	}
	if rule.Price().Amount() != 15000 {
		t.Errorf("Expected Price 15000, got %d", rule.Price().Amount())
	}
}

func TestPriceRuleReconstitute(t *testing.T) {
	startDate, _ := time.Parse("2006-01-02", "2024-07-01")
	endDate, _ := time.Parse("2006-01-02", "2024-07-31")
	dateRange, _ := vo.NewDateRange(startDate, endDate)
	price := vo.NewMoney(20000, "EUR")
	createdAt := time.Now()
	
	rule := ReconstitutePriceRule("rule-456", "unit-type-789", dateRange, price, createdAt)
	
	if rule.ID() != "rule-456" {
		t.Errorf("Expected ID 'rule-456', got '%s'", rule.ID())
	}
	if rule.UnitTypeID() != "unit-type-789" {
		t.Errorf("Expected UnitTypeID 'unit-type-789', got '%s'", rule.UnitTypeID())
	}
	if rule.Price().Currency() != "EUR" {
		t.Errorf("Expected Currency 'EUR', got '%s'", rule.Price().Currency())
	}
}

func TestPriceRuleGetters(t *testing.T) {
	startDate, _ := time.Parse("2006-01-02", "2024-08-01")
	endDate, _ := time.Parse("2006-01-02", "2024-08-15")
	dateRange, _ := vo.NewDateRange(startDate, endDate)
	price := vo.NewMoney(25000, "GBP")
	
	rule := NewPriceRule("unit-type-abc", dateRange, price)
	
	if rule.ID() == "" {
		t.Error("Expected non-empty ID")
	}
	if rule.UnitTypeID() != "unit-type-abc" {
		t.Errorf("Expected UnitTypeID 'unit-type-abc', got '%s'", rule.UnitTypeID())
	}
	if rule.DateRange().Start() != startDate {
		t.Errorf("Expected Start %v, got %v", startDate, rule.DateRange().Start())
	}
	if rule.DateRange().End() != endDate {
		t.Errorf("Expected End %v, got %v", endDate, rule.DateRange().End())
	}
	if rule.Price().Amount() != 25000 {
		t.Errorf("Expected Price 25000, got %d", rule.Price().Amount())
	}
}
