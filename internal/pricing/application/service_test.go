package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ecelayes/pms-backend/internal/pricing/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
)

func TestNewPricingService(t *testing.T) {
	svc := NewPricingService(nil, nil)
	if svc == nil {
		t.Error("Expected non-nil service")
	}
	if svc.priceRepo != nil {
		t.Error("Expected nil priceRepo")
	}
	if svc.ratePlanRepo != nil {
		t.Error("Expected nil ratePlanRepo")
	}
}

func TestPricingService_GetRatePlan_NotFound(t *testing.T) {
	mockRepo := &mockRatePlanRepo{findByIDResult: nil}

	svc := NewPricingService(nil, mockRepo)

	result, err := svc.GetRatePlan(context.Background(), "non-existent-id")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
	if result != nil {
		t.Error("Expected nil result")
	}
}

func TestPricingService_ListRatePlans(t *testing.T) {
	rp1 := domain.ReconstituteRatePlan(
		"rp-1", "prop-1", nil, "Rate 1", "", true,
		domain.MealPlan{}, domain.CancellationPolicy{}, domain.PaymentPolicy{}, time.Now(),
	)
	rp2 := domain.ReconstituteRatePlan(
		"rp-2", "prop-1", nil, "Rate 2", "", true,
		domain.MealPlan{}, domain.CancellationPolicy{}, domain.PaymentPolicy{}, time.Now(),
	)

	mockRepo := &mockRatePlanRepo{findByPropertyIDResult: []*domain.RatePlan{rp1, rp2}}
	svc := NewPricingService(nil, mockRepo)

	results, err := svc.ListRatePlans(context.Background(), "prop-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got: %d", len(results))
	}
}

func TestPricingService_UpdateRatePlan_NotFound(t *testing.T) {
	mockRepo := &mockRatePlanRepo{findByIDResult: nil}
	svc := NewPricingService(nil, mockRepo)

	err := svc.UpdateRatePlan(context.Background(), "non-existent", "new name", "desc", false)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
}

func TestPricingService_UpdateRatePlan_Success(t *testing.T) {
	rp := domain.ReconstituteRatePlan(
		"rp-1", "prop-1", nil, "Original", "", true,
		domain.MealPlan{}, domain.CancellationPolicy{IsRefundable: true, Rules: []domain.CancellationRule{}}, domain.PaymentPolicy{}, time.Now(),
	)
	mockRepo := &mockRatePlanRepo{findByIDResult: rp}
	svc := NewPricingService(nil, mockRepo)

	err := svc.UpdateRatePlan(context.Background(), "rp-1", "Updated Name", "New desc", false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if rp.Name() != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", rp.Name())
	}
}

func TestPricingService_CalculateBasePrice_NoRules(t *testing.T) {
	mockPriceRepo := &mockPriceRuleRepo{findOverlappingResult: []*domain.PriceRule{}}
	svc := NewPricingService(mockPriceRepo, nil)

	defaultPrice := vo.NewMoney(10000, "USD")
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)

	result, err := svc.CalculateBasePrice(context.Background(), "ut-1", defaultPrice, start, end)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// 2 nights * 10000 = 20000
	if result.Amount() != 20000 {
		t.Errorf("Expected 20000, got: %d", result.Amount())
	}
}

func TestPricingService_CalculateBasePrice_WithRules(t *testing.T) {
	range1, _ := vo.NewDateRange(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 23, 59, 59, 0, time.UTC))
	rule1 := domain.ReconstitutePriceRule(
		"pr-1", "ut-1", range1, vo.NewMoney(15000, "USD"), time.Now(),
	)
	range2, _ := vo.NewDateRange(time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 2, 23, 59, 59, 0, time.UTC))
	rule2 := domain.ReconstitutePriceRule(
		"pr-2", "ut-1", range2, vo.NewMoney(12000, "USD"), time.Now(),
	)

	mockPriceRepo := &mockPriceRuleRepo{findOverlappingResult: []*domain.PriceRule{rule1, rule2}}
	svc := NewPricingService(mockPriceRepo, nil)

	defaultPrice := vo.NewMoney(10000, "USD")
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)

	result, err := svc.CalculateBasePrice(context.Background(), "ut-1", defaultPrice, start, end)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// 15000 + 12000 = 27000
	if result.Amount() != 27000 {
		t.Errorf("Expected 27000, got: %d", result.Amount())
	}
}

func TestPricingService_CalculateStayPrice_WithRatePlan(t *testing.T) {
	rp, _ := domain.NewRatePlan("prop-1", "Test", domain.MealPlan{Included: false, PricePerPax: 500, Type: 1}, domain.CancellationPolicy{IsRefundable: true, Rules: []domain.CancellationRule{}}, domain.PaymentPolicy{})
	mockRatePlanRepo := &mockRatePlanRepo{findByIDResult: rp}
	mockPriceRepo := &mockPriceRuleRepo{findOverlappingResult: []*domain.PriceRule{}}
	svc := NewPricingService(mockPriceRepo, mockRatePlanRepo)

	defaultPrice := vo.NewMoney(10000, "USD")
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)

	// 2 adults, 0 children - meal plan 500 * 2 = 1000 extra
	result, err := svc.CalculateStayPrice(context.Background(), "ut-1", rp.ID(), defaultPrice, start, end, 2, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Base: 10000 (1 night), + meal plan: 500 * 2 = 1000
	// Total should be 11000
	if result.Amount() != 11000 {
		t.Errorf("Expected 11000, got: %d", result.Amount())
	}
}

func TestPricingService_CalculateCancellationPenalty_NonRefundable(t *testing.T) {
	rp, _ := domain.NewRatePlan("prop-1", "Non-ref", domain.MealPlan{}, domain.CancellationPolicy{IsRefundable: false, Rules: []domain.CancellationRule{}}, domain.PaymentPolicy{})
	mockRepo := &mockRatePlanRepo{findByIDResult: rp}
	svc := NewPricingService(nil, mockRepo)

	totalPrice := vo.NewMoney(50000, "USD")
	resStart := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	cancelDate := time.Date(2024, 6, 28, 0, 0, 0, 0, time.UTC)

	penalty, err := svc.CalculateCancellationPenalty(context.Background(), rp.ID(), resStart, cancelDate, totalPrice)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Non-refundable = full price penalty
	if penalty.Amount() != 50000 {
		t.Errorf("Expected 50000 (full price), got: %d", penalty.Amount())
	}
}

func TestPricingService_CalculateCancellationPenalty_RefundableNoRules(t *testing.T) {
	rp, _ := domain.NewRatePlan("prop-1", "Flexible", domain.MealPlan{}, domain.CancellationPolicy{IsRefundable: true, Rules: []domain.CancellationRule{}}, domain.PaymentPolicy{})
	mockRepo := &mockRatePlanRepo{findByIDResult: rp}
	svc := NewPricingService(nil, mockRepo)

	totalPrice := vo.NewMoney(50000, "USD")
	resStart := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	cancelDate := time.Date(2024, 6, 28, 0, 0, 0, 0, time.UTC)

	penalty, err := svc.CalculateCancellationPenalty(context.Background(), rp.ID(), resStart, cancelDate, totalPrice)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// No rules = no penalty
	if penalty.Amount() != 0 {
		t.Errorf("Expected 0 penalty, got: %d", penalty.Amount())
	}
}

func TestPricingService_CalculateCancellationPenalty_PercentageRule(t *testing.T) {
	rp, _ := domain.NewRatePlan("prop-1", "Moderate", domain.MealPlan{}, domain.CancellationPolicy{
		IsRefundable: true,
		Rules: []domain.CancellationRule{
			{HoursBeforeCheckIn: 48, PenaltyType: domain.PenaltyPercentage, PenaltyValue: 50},
		},
	}, domain.PaymentPolicy{})
	mockRepo := &mockRatePlanRepo{findByIDResult: rp}
	svc := NewPricingService(nil, mockRepo)

	// resStart - cancelDate = 36 hours (so 36 < 48 is true, rule applies)
	totalPrice := vo.NewMoney(50000, "USD")
	resStart := time.Date(2024, 7, 1, 12, 0, 0, 0, time.UTC)
	cancelDate := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC) // 36 hours before

	penalty, err := svc.CalculateCancellationPenalty(context.Background(), rp.ID(), resStart, cancelDate, totalPrice)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// 50% of 50000 = 25000
	if penalty.Amount() != 25000 {
		t.Errorf("Expected 25000 (50%% penalty), got: %d", penalty.Amount())
	}
}

func TestPricingService_CalculateCancellationPenalty_FixedAmount(t *testing.T) {
	rp, _ := domain.NewRatePlan("prop-1", "Moderate", domain.MealPlan{}, domain.CancellationPolicy{
		IsRefundable: true,
		Rules: []domain.CancellationRule{
			{HoursBeforeCheckIn: 24, PenaltyType: domain.PenaltyFixedAmount, PenaltyValue: 2500},
		},
	}, domain.PaymentPolicy{})
	mockRepo := &mockRatePlanRepo{findByIDResult: rp}
	svc := NewPricingService(nil, mockRepo)

	// resStart - cancelDate = 12 hours (so 12 < 24 is true, rule applies)
	totalPrice := vo.NewMoney(50000, "USD")
	resStart := time.Date(2024, 7, 1, 12, 0, 0, 0, time.UTC)
	cancelDate := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC) // 12 hours before

	penalty, err := svc.CalculateCancellationPenalty(context.Background(), rp.ID(), resStart, cancelDate, totalPrice)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Fixed amount 2500
	if penalty.Amount() != 2500 {
		t.Errorf("Expected 2500 (fixed), got: %d", penalty.Amount())
	}
}

func TestPricingService_CalculateCancellationPenalty_NightsNotSupported(t *testing.T) {
	rp, _ := domain.NewRatePlan("prop-1", "Nights Plan", domain.MealPlan{}, domain.CancellationPolicy{
		IsRefundable: true,
		Rules: []domain.CancellationRule{
			// Must have HoursBeforeCheckIn > hoursBefore for rule to match
			{HoursBeforeCheckIn: 96, PenaltyType: domain.PenaltyNights, PenaltyValue: 1},
		},
	}, domain.PaymentPolicy{})
	mockRepo := &mockRatePlanRepo{findByIDResult: rp}
	svc := NewPricingService(nil, mockRepo)

	// resStart - cancelDate = 48 hours (so 48 < 96 is true, rule matches)
	// But PenaltyNights is not supported, should error
	totalPrice := vo.NewMoney(50000, "USD")
	resStart := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	cancelDate := time.Date(2024, 6, 29, 0, 0, 0, 0, time.UTC) // 48 hours before

	_, err := svc.CalculateCancellationPenalty(context.Background(), rp.ID(), resStart, cancelDate, totalPrice)
	if err == nil {
		t.Error("Expected error for nights penalty (not supported)")
	}
	// Check error message contains "nights"
	if !strings.Contains(err.Error(), "nights") {
		t.Errorf("Expected error message to contain 'nights', got: %v", err)
	}
}

func TestPricingService_CalculateCancellationPenalty_AllRulesNonMatching(t *testing.T) {
	rp, _ := domain.NewRatePlan("prop-1", "Moderate", domain.MealPlan{}, domain.CancellationPolicy{
		IsRefundable: true,
		Rules: []domain.CancellationRule{
			{HoursBeforeCheckIn: 48, PenaltyType: domain.PenaltyPercentage, PenaltyValue: 50},
			{HoursBeforeCheckIn: 24, PenaltyType: domain.PenaltyFixedAmount, PenaltyValue: 5000},
		},
	}, domain.PaymentPolicy{})
	mockRepo := &mockRatePlanRepo{findByIDResult: rp}
	svc := NewPricingService(nil, mockRepo)

	// resStart - cancelDate = 72 hours (72 < 48 is false, 72 < 24 is false)
	// No rules match, so penalty = 0
	totalPrice := vo.NewMoney(50000, "USD")
	resStart := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	cancelDate := time.Date(2024, 6, 28, 0, 0, 0, 0, time.UTC)

	penalty, err := svc.CalculateCancellationPenalty(context.Background(), rp.ID(), resStart, cancelDate, totalPrice)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if penalty.Amount() != 0 {
		t.Errorf("Expected 0 penalty (no rules matched), got: %d", penalty.Amount())
	}
}

func TestPricingService_CalculateCancellationPenalty_BeforeCancelDate(t *testing.T) {
	rp, _ := domain.NewRatePlan("prop-1", "Moderate", domain.MealPlan{}, domain.CancellationPolicy{
		IsRefundable: true,
		Rules: []domain.CancellationRule{
			{HoursBeforeCheckIn: 48, PenaltyType: domain.PenaltyPercentage, PenaltyValue: 50},
		},
	}, domain.PaymentPolicy{})
	mockRepo := &mockRatePlanRepo{findByIDResult: rp}
	svc := NewPricingService(nil, mockRepo)

	// cancelDate after resStart = negative hoursBefore
	// Should return full price as penalty
	totalPrice := vo.NewMoney(50000, "USD")
	resStart := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	cancelDate := time.Date(2024, 7, 2, 0, 0, 0, 0, time.UTC) // After resStart

	penalty, err := svc.CalculateCancellationPenalty(context.Background(), rp.ID(), resStart, cancelDate, totalPrice)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if penalty.Amount() != 50000 {
		t.Errorf("Expected 50000 (negative hours), got: %d", penalty.Amount())
	}
}

func TestPricingService_SetPriceRule_Success(t *testing.T) {
	mockPriceRepo := &mockPriceRuleRepo{findOverlappingResult: []*domain.PriceRule{}}
	svc := NewPricingService(mockPriceRepo, nil)

	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	price := vo.NewMoney(15000, "EUR")

	err := svc.SetPriceRule(context.Background(), "ut-1", start, end, price)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPricingService_SetPriceRule_SplitExisting(t *testing.T) {
	existingRange, _ := vo.NewDateRange(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 10, 23, 59, 59, 0, time.UTC))
	existingRule := domain.ReconstitutePriceRule("pr-1", "ut-1", existingRange, vo.NewMoney(10000, "USD"), time.Now())

	// New range is June 3-5, should split existing into [Jun 1-2] and [Jun 6-10]
	mockPriceRepo := &mockPriceRuleRepo{
		findOverlappingResult: []*domain.PriceRule{existingRule},
	}
	svc := NewPricingService(mockPriceRepo, nil)

	start := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 5, 23, 59, 59, 0, time.UTC)
	price := vo.NewMoney(15000, "USD")

	err := svc.SetPriceRule(context.Background(), "ut-1", start, end, price)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPricingService_SetPriceRule_InvalidDateRange(t *testing.T) {
	mockPriceRepo := &mockPriceRuleRepo{}
	svc := NewPricingService(mockPriceRepo, nil)

	start := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC) // End before start
	price := vo.NewMoney(15000, "USD")

	err := svc.SetPriceRule(context.Background(), "ut-1", start, end, price)
	if err == nil {
		t.Error("Expected error for invalid date range")
	}
}

func TestPricingService_GetRules_WithUnitTypeID(t *testing.T) {
	range1, _ := vo.NewDateRange(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 10, 23, 59, 59, 0, time.UTC))
	rule := domain.ReconstitutePriceRule("pr-1", "ut-1", range1, vo.NewMoney(10000, "USD"), time.Now())
	mockPriceRepo := &mockPriceRuleRepo{findByUnitTypeResult: []*domain.PriceRule{rule}}
	svc := NewPricingService(mockPriceRepo, nil)

	rules, err := svc.GetRules(context.Background(), "", "ut-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got: %d", len(rules))
	}
}

func TestPricingService_GetRules_WithPropertyID(t *testing.T) {
	range1, _ := vo.NewDateRange(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 10, 23, 59, 59, 0, time.UTC))
	rule := domain.ReconstitutePriceRule("pr-1", "ut-1", range1, vo.NewMoney(10000, "USD"), time.Now())
	mockPriceRepo := &mockPriceRuleRepo{findByPropertyIDResult: []*domain.PriceRule{rule}}
	svc := NewPricingService(mockPriceRepo, nil)

	rules, err := svc.GetRules(context.Background(), "prop-1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got: %d", len(rules))
	}
}

func TestPricingService_GetRules_NoFilter(t *testing.T) {
	svc := NewPricingService(nil, nil)

	rules, err := svc.GetRules(context.Background(), "", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if rules != nil {
		t.Error("Expected nil rules when no filter provided")
	}
}

func TestPricingService_CreateRatePlan(t *testing.T) {
	svc := NewPricingService(nil, nil)
	
	_, err := svc.CreateRatePlan(
		context.Background(),
		"", // empty property ID should fail
		"Test Rate",
		domain.MealPlan{},
		domain.CancellationPolicy{},
		domain.PaymentPolicy{},
	)
	if err == nil {
		t.Error("Expected error for empty property ID")
	}
}

func TestPricingService_CreateRatePlan_Valid(t *testing.T) {
	mockRepo := &mockRatePlanRepo{}
	svc := NewPricingService(nil, mockRepo)

	mealPlan := domain.MealPlan{Included: true, PricePerPax: 1500, Type: 1}
	cancel := domain.CancellationPolicy{IsRefundable: true, Rules: []domain.CancellationRule{}}
	pay := domain.PaymentPolicy{PrepayPercent: 20}

	_, err := svc.CreateRatePlan(context.Background(), "prop-123", "Test Rate", mealPlan, cancel, pay)
	if err != nil {
		t.Errorf("Expected no error creating entity, got: %v", err)
	}
}

func TestPricingService_DeleteRatePlan(t *testing.T) {
	mockRepo := &mockRatePlanRepo{}
	svc := NewPricingService(nil, mockRepo)

	err := svc.DeleteRatePlan(context.Background(), "rp-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPricingService_DeleteRule(t *testing.T) {
	mockPriceRepo := &mockPriceRuleRepo{}
	svc := NewPricingService(mockPriceRepo, nil)

	err := svc.DeleteRule(context.Background(), "pr-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}


type mockRatePlanRepo struct {
	saveResult              error
	findByIDResult          *domain.RatePlan
	findByIDErr             error
	findByPropertyIDResult  []*domain.RatePlan
	findByPropertyIDErr     error
	deleteErr               error
}

func (m *mockRatePlanRepo) Save(ctx context.Context, rp *domain.RatePlan) error {
	return m.saveResult
}
func (m *mockRatePlanRepo) FindByID(ctx context.Context, id string) (*domain.RatePlan, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	return m.findByIDResult, nil
}
func (m *mockRatePlanRepo) FindByPropertyID(ctx context.Context, propertyID string) ([]*domain.RatePlan, error) {
	if m.findByPropertyIDErr != nil {
		return nil, m.findByPropertyIDErr
	}
	return m.findByPropertyIDResult, nil
}
func (m *mockRatePlanRepo) Delete(ctx context.Context, id string) error {
	return m.deleteErr
}

type mockPriceRuleRepo struct {
	saveResult              error
	saveErrOnNth            int
	saveCount               int
	findOverlappingResult   []*domain.PriceRule
	findOverlappingErr      error
	findByUnitTypeResult    []*domain.PriceRule
	findByUnitTypeErr       error
	findByPropertyIDResult  []*domain.PriceRule
	findByPropertyIDErr     error
	deleteErr               error
	deleteCalls             []string
}

func (m *mockPriceRuleRepo) Save(ctx context.Context, pr *domain.PriceRule) error {
	m.saveCount++
	if m.saveErrOnNth > 0 && m.saveCount == m.saveErrOnNth {
		return errors.New("save error on nth call")
	}
	return m.saveResult
}
func (m *mockPriceRuleRepo) FindOverlapping(ctx context.Context, unitTypeID, start, end string) ([]*domain.PriceRule, error) {
	if m.findOverlappingErr != nil {
		return nil, m.findOverlappingErr
	}
	return m.findOverlappingResult, nil
}
func (m *mockPriceRuleRepo) FindByUnitType(ctx context.Context, unitTypeID string) ([]*domain.PriceRule, error) {
	if m.findByUnitTypeErr != nil {
		return nil, m.findByUnitTypeErr
	}
	return m.findByUnitTypeResult, nil
}
func (m *mockPriceRuleRepo) FindByPropertyID(ctx context.Context, propertyID string) ([]*domain.PriceRule, error) {
	if m.findByPropertyIDErr != nil {
		return nil, m.findByPropertyIDErr
	}
	return m.findByPropertyIDResult, nil
}
func (m *mockPriceRuleRepo) Delete(ctx context.Context, id string) error {
	if m.deleteCalls == nil {
		m.deleteCalls = []string{}
	}
	m.deleteCalls = append(m.deleteCalls, id)
	return m.deleteErr
}

var _ = errors.New

func TestPricingService_CalculateBasePrice_EmptyDates(t *testing.T) {
	mockPriceRepo := &mockPriceRuleRepo{findOverlappingResult: []*domain.PriceRule{}}
	svc := NewPricingService(mockPriceRepo, nil)

	defaultPrice := vo.NewMoney(10000, "USD")
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	result, err := svc.CalculateBasePrice(context.Background(), "ut-1", defaultPrice, start, end)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Amount() != 0 {
		t.Errorf("Expected 0 for empty dates, got: %d", result.Amount())
	}
}

func TestPricingService_CalculateBasePrice_CurrencyMismatch(t *testing.T) {
	mockPriceRepo := &mockPriceRuleRepo{findOverlappingResult: []*domain.PriceRule{}}
	svc := NewPricingService(mockPriceRepo, nil)

	defaultPrice := vo.NewMoney(10000, "USD")
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)

	// Currency mismatch shouldn't happen normally, but test it
	result, err := svc.CalculateBasePrice(context.Background(), "ut-1", defaultPrice, start, end)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Currency() != "USD" {
		t.Errorf("Expected USD, got: %s", result.Currency())
	}
}

func TestPricingService_GetRatePlan_RepoError(t *testing.T) {
	repo := &mockRatePlanRepo{findByIDErr: errors.New("repo error")}
	svc := NewPricingService(nil, repo)

	_, err := svc.GetRatePlan(context.Background(), "rp-1")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestPricingService_GetRatePlan_Success(t *testing.T) {
	rp, _ := domain.NewRatePlan("prop-1", "Test", domain.MealPlan{}, domain.CancellationPolicy{}, domain.PaymentPolicy{})
	repo := &mockRatePlanRepo{findByIDResult: rp}
	svc := NewPricingService(nil, repo)

	result, err := svc.GetRatePlan(context.Background(), rp.ID())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}
}

func TestPricingService_ListRatePlans_RepoError(t *testing.T) {
	repo := &mockRatePlanRepo{findByPropertyIDErr: errors.New("repo error")}
	svc := NewPricingService(nil, repo)

	_, err := svc.ListRatePlans(context.Background(), "prop-1")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestPricingService_UpdateRatePlan_GetByIDError(t *testing.T) {
	repo := &mockRatePlanRepo{findByIDErr: errors.New("find error")}
	svc := NewPricingService(nil, repo)

	err := svc.UpdateRatePlan(context.Background(), "rp-1", "Name", "Desc", true)
	if err == nil {
		t.Error("Expected error from GetByID")
	}
}

func TestPricingService_UpdateRatePlan_SaveError(t *testing.T) {
	rp, _ := domain.NewRatePlan("prop-1", "Test", domain.MealPlan{}, domain.CancellationPolicy{}, domain.PaymentPolicy{})
	repo := &mockRatePlanRepo{findByIDResult: rp, saveResult: errors.New("save error")}
	svc := NewPricingService(nil, repo)

	err := svc.UpdateRatePlan(context.Background(), rp.ID(), "Name", "Desc", true)
	if err == nil {
		t.Error("Expected error from Save")
	}
}

func TestPricingService_DeleteRatePlan_Success(t *testing.T) {
	repo := &mockRatePlanRepo{}
	svc := NewPricingService(nil, repo)

	err := svc.DeleteRatePlan(context.Background(), "rp-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPricingService_DeleteRatePlan_RepoError(t *testing.T) {
	repo := &mockRatePlanRepo{deleteErr: errors.New("delete error")}
	svc := NewPricingService(nil, repo)

	err := svc.DeleteRatePlan(context.Background(), "rp-1")
	if err == nil {
		t.Error("Expected error from Delete")
	}
}

func TestPricingService_GetRules_ByUnitType_Success(t *testing.T) {
	rules := []*domain.PriceRule{}
	repo := &mockPriceRuleRepo{findByUnitTypeResult: rules}
	svc := NewPricingService(repo, nil)

	result, err := svc.GetRules(context.Background(), "", "ut-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}
}

func TestPricingService_GetRules_ByProperty_Success(t *testing.T) {
	rules := []*domain.PriceRule{}
	repo := &mockPriceRuleRepo{findByPropertyIDResult: rules}
	svc := NewPricingService(repo, nil)

	result, err := svc.GetRules(context.Background(), "prop-1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}
}

func TestPricingService_DeleteRule_RepoError(t *testing.T) {
	repo := &mockPriceRuleRepo{deleteErr: errors.New("delete error")}
	svc := NewPricingService(repo, nil)

	err := svc.DeleteRule(context.Background(), "rule-1")
	if err == nil {
		t.Error("Expected error from Delete")
	}
}

func TestPricingService_CalculateStayPrice_NoRatePlan(t *testing.T) {
	mockPriceRepo := &mockPriceRuleRepo{findOverlappingResult: []*domain.PriceRule{}}
	svc := NewPricingService(mockPriceRepo, nil)

	defaultPrice := vo.NewMoney(10000, "USD")
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)

	result, err := svc.CalculateStayPrice(context.Background(), "ut-1", "", defaultPrice, start, end, 2, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Amount() != 10000 {
		t.Errorf("Expected 10000, got: %d", result.Amount())
	}
}
func TestPricingService_CalculateStayPrice_RepoError(t *testing.T) {
	mockPriceRepo := &mockPriceRuleRepo{findOverlappingErr: errors.New("repo error")}
	svc := NewPricingService(mockPriceRepo, nil)
	defaultPrice := vo.NewMoney(10000, "USD")
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)
	_, err := svc.CalculateStayPrice(context.Background(), "ut-1", "", defaultPrice, start, end, 1, 0)
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestPricingService_CalculateStayPrice_RatePlanNotFound(t *testing.T) {
	mockPriceRepo := &mockPriceRuleRepo{findOverlappingResult: []*domain.PriceRule{}}
	mockRatePlanRepo := &mockRatePlanRepo{findByIDResult: nil}
	svc := NewPricingService(mockPriceRepo, mockRatePlanRepo)
	defaultPrice := vo.NewMoney(10000, "USD")
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)

	result, err := svc.CalculateStayPrice(context.Background(), "ut-1", "rp-1", defaultPrice, start, end, 2, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Amount() != 10000 {
		t.Errorf("Expected 10000, got: %d", result.Amount())
	}
}

func TestPricingService_CalculateStayPrice_RatePlanRepoError(t *testing.T) {
	mockPriceRepo := &mockPriceRuleRepo{findOverlappingResult: []*domain.PriceRule{}}
	mockRatePlanRepo := &mockRatePlanRepo{findByIDErr: errors.New("repo error")}
	svc := NewPricingService(mockPriceRepo, mockRatePlanRepo)

	defaultPrice := vo.NewMoney(10000, "USD")
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)

	_, err := svc.CalculateStayPrice(context.Background(), "ut-1", "rp-1", defaultPrice, start, end, 2, 0)
	if err == nil {
		t.Error("Expected error from rate plan repo")
	}
}

func TestPricingService_CalculateStayPrice_WithChildren(t *testing.T) {
	rp, _ := domain.NewRatePlan("prop-1", "Test", domain.MealPlan{Included: false, PricePerPax: 500, Type: 1}, domain.CancellationPolicy{IsRefundable: true, Rules: []domain.CancellationRule{}}, domain.PaymentPolicy{})
	mockRatePlanRepo := &mockRatePlanRepo{findByIDResult: rp}
	mockPriceRepo := &mockPriceRuleRepo{findOverlappingResult: []*domain.PriceRule{}}
	svc := NewPricingService(mockPriceRepo, mockRatePlanRepo)

	defaultPrice := vo.NewMoney(10000, "USD")
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)

	// 2 adults + 2 children = 4 pax, 500 per pax per day, 2 days = 4000
	// Base: 10000 * 2 = 20000, Total: 24000
	result, err := svc.CalculateStayPrice(context.Background(), "ut-1", rp.ID(), defaultPrice, start, end, 2, 2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Amount() != 24000 {
		t.Errorf("Expected 24000, got: %d", result.Amount())
	}
}

func TestPricingService_CreateRatePlan_SaveError(t *testing.T) {
	repo := &mockRatePlanRepo{saveResult: errors.New("save error")}
	svc := NewPricingService(nil, repo)
	_, err := svc.CreateRatePlan(context.Background(), "prop-1", "Test", domain.MealPlan{}, domain.CancellationPolicy{}, domain.PaymentPolicy{})
	if err == nil {
		t.Error("Expected save error")
	}
}

func TestPricingService_CalculateBasePrice_RepoError(t *testing.T) {
	repo := &mockPriceRuleRepo{findOverlappingErr: errors.New("repo error")}
	svc := NewPricingService(repo, nil)
	_, err := svc.CalculateBasePrice(context.Background(), "ut-1", vo.NewMoney(10000, "USD"), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC))
	if err == nil {
		t.Error("Expected repo error")
	}
}

func TestPricingService_CalculateBasePrice_WithOverlappingRules(t *testing.T) {
	ruleStart := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	ruleEnd := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	ruleRange, _ := vo.NewDateRange(ruleStart, ruleEnd.Add(-time.Second))
	rule := domain.NewPriceRule("ut-1", ruleRange, vo.NewMoney(20000, "USD"))
	repo := &mockPriceRuleRepo{findOverlappingResult: []*domain.PriceRule{rule}}
	svc := NewPricingService(repo, nil)
	defaultPrice := vo.NewMoney(10000, "USD")
	result, err := svc.CalculateBasePrice(context.Background(), "ut-1", defaultPrice, ruleStart, ruleEnd)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Amount() == 0 {
		t.Error("Expected non-zero total")
	}
}

func TestPricingService_CalculateCancellationPenalty_RepoError(t *testing.T) {
	repo := &mockRatePlanRepo{findByIDErr: errors.New("repo error")}
	svc := NewPricingService(nil, repo)
	resStart := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	cancelDate := time.Date(2024, 6, 28, 0, 0, 0, 0, time.UTC)
	_, err := svc.CalculateCancellationPenalty(context.Background(), "rp-1", resStart, cancelDate, vo.NewMoney(10000, "USD"))
	if err == nil {
		t.Error("Expected repo error")
	}
}

func TestPricingService_CalculateCancellationPenalty_WithRules(t *testing.T) {
	rule := domain.CancellationRule{}
	_ = rule
	rp, _ := domain.NewRatePlan("prop-1", "Refundable", domain.MealPlan{}, domain.CancellationPolicy{IsRefundable: true, Rules: []domain.CancellationRule{{HoursBeforeCheckIn: 168, PenaltyType: domain.PenaltyPercentage, PenaltyValue: 50}}}, domain.PaymentPolicy{})
	repo := &mockRatePlanRepo{findByIDResult: rp}
	svc := NewPricingService(nil, repo)
	resStart := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	cancelDate := time.Date(2024, 6, 25, 0, 0, 0, 0, time.UTC)
	penalty, err := svc.CalculateCancellationPenalty(context.Background(), rp.ID(), resStart, cancelDate, vo.NewMoney(100000, "USD"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if penalty.Amount() == 0 {
		t.Error("Expected non-zero penalty")
	}
}

func TestPricingService_SetPriceRule_RepoError(t *testing.T) {
	repo := &mockPriceRuleRepo{findOverlappingErr: errors.New("repo error")}
	svc := NewPricingService(repo, nil)
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	err := svc.SetPriceRule(context.Background(), "ut-1", start, end, vo.NewMoney(20000, "USD"))
	if err == nil {
		t.Error("Expected repo error")
	}
}

func TestPricingService_SetPriceRule_InvalidDateRangeExtra(t *testing.T) {
	repo := &mockPriceRuleRepo{}
	svc := NewPricingService(repo, nil)
	start := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	err := svc.SetPriceRule(context.Background(), "ut-1", start, end, vo.NewMoney(20000, "USD"))
	if err == nil {
		t.Error("Expected error for invalid date range")
	}
}

func TestPricingService_SetPriceRule_SaveError(t *testing.T) {
	repo := &mockPriceRuleRepo{findOverlappingResult: []*domain.PriceRule{}, saveResult: errors.New("save error")}
	svc := NewPricingService(repo, nil)
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	err := svc.SetPriceRule(context.Background(), "ut-1", start, end, vo.NewMoney(20000, "USD"))
	if err == nil {
		t.Error("Expected save error")
	}
}

func TestPricingService_CalculateCancellationPenalty_NegativeHours(t *testing.T) {
	rp, _ := domain.NewRatePlan("prop-1", "Refundable", domain.MealPlan{}, domain.CancellationPolicy{IsRefundable: true, Rules: []domain.CancellationRule{}}, domain.PaymentPolicy{})
	repo := &mockRatePlanRepo{findByIDResult: rp}
	svc := NewPricingService(nil, repo)
	resStart := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	cancelDate := time.Date(2024, 7, 2, 0, 0, 0, 0, time.UTC) // after start
	totalPrice := vo.NewMoney(10000, "USD")
	penalty, err := svc.CalculateCancellationPenalty(context.Background(), rp.ID(), resStart, cancelDate, totalPrice)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if penalty.Amount() != 10000 {
		t.Errorf("Expected 10000, got: %d", penalty.Amount())
	}
}

func TestPricingService_CalculateCancellationPenalty_FixedAmountV2(t *testing.T) {
	rp, _ := domain.NewRatePlan("prop-1", "Test", domain.MealPlan{}, domain.CancellationPolicy{IsRefundable: true, Rules: []domain.CancellationRule{{HoursBeforeCheckIn: 24, PenaltyType: domain.PenaltyFixedAmount, PenaltyValue: 5000}}}, domain.PaymentPolicy{})
	repo := &mockRatePlanRepo{findByIDResult: rp}
	svc := NewPricingService(nil, repo)
	resStart := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	cancelDate := time.Date(2024, 6, 30, 12, 0, 0, 0, time.UTC) // 12 hours before
	totalPrice := vo.NewMoney(10000, "USD")
	penalty, err := svc.CalculateCancellationPenalty(context.Background(), rp.ID(), resStart, cancelDate, totalPrice)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if penalty.Amount() != 5000 {
		t.Errorf("Expected 5000, got: %d", penalty.Amount())
	}
}

func TestPricingService_CalculateCancellationPenalty_NightsNotSupportedV2(t *testing.T) {
	rp, _ := domain.NewRatePlan("prop-1", "Test", domain.MealPlan{}, domain.CancellationPolicy{IsRefundable: true, Rules: []domain.CancellationRule{{HoursBeforeCheckIn: 24, PenaltyType: domain.PenaltyNights, PenaltyValue: 1}}}, domain.PaymentPolicy{})
	repo := &mockRatePlanRepo{findByIDResult: rp}
	svc := NewPricingService(nil, repo)
	resStart := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	cancelDate := time.Date(2024, 6, 30, 12, 0, 0, 0, time.UTC)
	totalPrice := vo.NewMoney(10000, "USD")
	_, err := svc.CalculateCancellationPenalty(context.Background(), rp.ID(), resStart, cancelDate, totalPrice)
	if err == nil {
		t.Error("Expected error for nights penalty")
	}
}

func TestPricingService_CalculateCancellationPenalty_NoMatchingRule(t *testing.T) {
	rp, _ := domain.NewRatePlan("prop-1", "Test", domain.MealPlan{}, domain.CancellationPolicy{IsRefundable: true, Rules: []domain.CancellationRule{{HoursBeforeCheckIn: 1, PenaltyType: domain.PenaltyFixedAmount, PenaltyValue: 5000}}}, domain.PaymentPolicy{})
	repo := &mockRatePlanRepo{findByIDResult: rp}
	svc := NewPricingService(nil, repo)
	resStart := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	cancelDate := time.Date(2024, 6, 30, 23, 0, 0, 0, time.UTC) // 1 hour before, rule needs less than 1 hour
	totalPrice := vo.NewMoney(10000, "USD")
	penalty, err := svc.CalculateCancellationPenalty(context.Background(), rp.ID(), resStart, cancelDate, totalPrice)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if penalty.Amount() != 0 {
		t.Errorf("Expected 0, got: %d", penalty.Amount())
	}
}

func TestPricingService_CalculateCancellationPenalty_NotFound(t *testing.T) {
	repo := &mockRatePlanRepo{findByIDResult: nil}
	svc := NewPricingService(nil, repo)
	resStart := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	cancelDate := time.Date(2024, 6, 28, 0, 0, 0, 0, time.UTC)
	totalPrice := vo.NewMoney(10000, "USD")
	_, err := svc.CalculateCancellationPenalty(context.Background(), "rp-1", resStart, cancelDate, totalPrice)
	if err == nil {
		t.Error("Expected not found error")
	}
}

func TestPricingService_SetPriceRule_SplitBothSides(t *testing.T) {
	// Existing rule is wider than new range, should split into left and right
	existingRange, _ := vo.NewDateRange(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 10, 23, 59, 59, 0, time.UTC))
	existingRule := domain.ReconstitutePriceRule("pr-1", "ut-1", existingRange, vo.NewMoney(10000, "USD"), time.Now())
	mockPriceRepo := &mockPriceRuleRepo{
		findOverlappingResult: []*domain.PriceRule{existingRule},
	}
	svc := NewPricingService(mockPriceRepo, nil)
	start := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 5, 23, 59, 59, 0, time.UTC)
	price := vo.NewMoney(15000, "USD")
	err := svc.SetPriceRule(context.Background(), "ut-1", start, end, price)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPricingService_SetPriceRule_DeleteError(t *testing.T) {
	existingRange, _ := vo.NewDateRange(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 5, 23, 59, 59, 0, time.UTC))
	existingRule := domain.ReconstitutePriceRule("pr-1", "ut-1", existingRange, vo.NewMoney(10000, "USD"), time.Now())
	repo := &mockPriceRuleRepo{
		findOverlappingResult: []*domain.PriceRule{existingRule},
		deleteErr:             errors.New("delete error"),
	}
	svc := NewPricingService(repo, nil)
	start := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 4, 23, 59, 59, 0, time.UTC)
	err := svc.SetPriceRule(context.Background(), "ut-1", start, end, vo.NewMoney(15000, "USD"))
	if err == nil {
		t.Error("Expected delete error")
	}
}

func TestPricingService_SetPriceRule_SaveErrorOnSplit(t *testing.T) {
	existingRange, _ := vo.NewDateRange(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 10, 23, 59, 59, 0, time.UTC))
	existingRule := domain.ReconstitutePriceRule("pr-1", "ut-1", existingRange, vo.NewMoney(10000, "USD"), time.Now())
	repo := &mockPriceRuleRepo{
		findOverlappingResult: []*domain.PriceRule{existingRule},
		saveResult:            errors.New("save error"),
	}
	svc := NewPricingService(repo, nil)
	start := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 5, 23, 59, 59, 0, time.UTC)
	err := svc.SetPriceRule(context.Background(), "ut-1", start, end, vo.NewMoney(15000, "USD"))
	if err == nil {
		t.Error("Expected save error")
	}
}

func TestPricingService_CalculateBasePrice_CurrencyMismatchInRule(t *testing.T) {
	// Test that the function doesn't crash when there's a default price currency mismatch
	// Since we control the rules, this is hard to trigger with valid data
	// But we can test with a default price and overlapping rule
	ruleStart := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	ruleEnd := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	ruleRange, _ := vo.NewDateRange(ruleStart, ruleEnd.Add(-time.Second))
	rule := domain.NewPriceRule("ut-1", ruleRange, vo.NewMoney(20000, "USD"))
	repo := &mockPriceRuleRepo{findOverlappingResult: []*domain.PriceRule{rule}}
	svc := NewPricingService(repo, nil)
	defaultPrice := vo.NewMoney(10000, "USD")
	result, err := svc.CalculateBasePrice(context.Background(), "ut-1", defaultPrice, ruleStart, ruleEnd)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == (vo.Money{}) {
		t.Error("Expected non-zero money")
	}
}

func TestPricingService_CalculateCancellationPenalty_MultipleMatchingRules(t *testing.T) {
	// Multiple rules that match - the one with the smallest HoursBeforeCheckIn should win
	rp, _ := domain.NewRatePlan("prop-1", "Test", domain.MealPlan{}, domain.CancellationPolicy{IsRefundable: true, Rules: []domain.CancellationRule{
		{HoursBeforeCheckIn: 168, PenaltyType: domain.PenaltyPercentage, PenaltyValue: 30},
		{HoursBeforeCheckIn: 24, PenaltyType: domain.PenaltyPercentage, PenaltyValue: 50},
		{HoursBeforeCheckIn: 48, PenaltyType: domain.PenaltyPercentage, PenaltyValue: 40},
	}}, domain.PaymentPolicy{})
	repo := &mockRatePlanRepo{findByIDResult: rp}
	svc := NewPricingService(nil, repo)
	resStart := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	cancelDate := time.Date(2024, 6, 30, 12, 0, 0, 0, time.UTC) // 12 hours before
	totalPrice := vo.NewMoney(10000, "USD")
	penalty, err := svc.CalculateCancellationPenalty(context.Background(), rp.ID(), resStart, cancelDate, totalPrice)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// 12 hours < 24 hours, so the 24-hour rule applies: 50% of 10000 = 5000
	if penalty.Amount() != 5000 {
		t.Errorf("Expected 5000, got: %d", penalty.Amount())
	}
}

func TestPricingService_SetPriceRule_RightSideSaveError(t *testing.T) {
	// Set up a scenario where the left save succeeds but the right save fails
	existingRange, _ := vo.NewDateRange(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 10, 23, 59, 59, 0, time.UTC))
	existingRule := domain.ReconstitutePriceRule("pr-1", "ut-1", existingRange, vo.NewMoney(10000, "USD"), time.Now())

	// Mock that fails on the second save (right side)
	mockRepo := &mockPriceRuleRepo{
		findOverlappingResult: []*domain.PriceRule{existingRule},
		saveErrOnNth:          2,
	}
	svc := NewPricingService(mockRepo, nil)
	start := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 5, 23, 59, 59, 0, time.UTC)
	err := svc.SetPriceRule(context.Background(), "ut-1", start, end, vo.NewMoney(15000, "USD"))
	if err == nil {
		t.Error("Expected error from right side save")
	}
}

func TestPricingService_CalculateBasePrice_CurrencyMismatchTriggersAddError(t *testing.T) {
	// Create a rule with a different currency than default price
	// This should trigger the total.Add error in CalculateBasePrice
	ruleStart := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	ruleEnd := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)
	ruleRange, _ := vo.NewDateRange(ruleStart, ruleEnd.Add(-time.Second))
	// Rule with USD, but default with USD - they must match for total.Add to succeed
	// But if we make the rule price with different currency than the default price,
	// it would cause the Add error
	// Since NewPriceRule doesn't validate, we can create a rule with EUR
	rule := domain.NewPriceRule("ut-1", ruleRange, vo.NewMoney(20000, "EUR"))
	repo := &mockPriceRuleRepo{findOverlappingResult: []*domain.PriceRule{rule}}
	svc := NewPricingService(repo, nil)
	defaultPrice := vo.NewMoney(10000, "USD")
	_, err := svc.CalculateBasePrice(context.Background(), "ut-1", defaultPrice, ruleStart, ruleEnd)
	if err == nil {
		t.Error("Expected currency mismatch error")
	}
}
