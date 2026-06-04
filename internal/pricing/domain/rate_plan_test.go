package domain

import (
	"testing"
	"time"

	"github.com/ecelayes/pms-backend/internal/shared/vo"
)

func TestNewRatePlan(t *testing.T) {
	mealPlan := MealPlan{Included: true, PricePerPax: 0, Type: 1}
	cancel := CancellationPolicy{IsRefundable: true, Rules: []CancellationRule{}}
	pay := PaymentPolicy{PrepayPercent: 20}
	
	rp, err := NewRatePlan("prop-123", "Standard Rate", mealPlan, cancel, pay)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rp.ID() == "" {
		t.Error("Expected non-empty ID")
	}
	if rp.PropertyID() != "prop-123" {
		t.Errorf("Expected PropertyID 'prop-123', got '%s'", rp.PropertyID())
	}
	if rp.Name() != "Standard Rate" {
		t.Errorf("Expected Name 'Standard Rate', got '%s'", rp.Name())
	}
	if !rp.Active() {
		t.Error("Expected Active to be true by default")
	}
}

func TestNewRatePlanEmptyPropertyID(t *testing.T) {
	mealPlan := MealPlan{Included: false, PricePerPax: 0, Type: 0}
	cancel := CancellationPolicy{IsRefundable: true, Rules: []CancellationRule{}}
	pay := PaymentPolicy{PrepayPercent: 0}
	
	rp, err := NewRatePlan("", "Standard Rate", mealPlan, cancel, pay)
	
	if err != ErrInvalidRatePlan {
		t.Errorf("Expected ErrInvalidRatePlan, got %v", err)
	}
	if rp != nil {
		t.Error("Expected nil RatePlan on invalid property ID")
	}
}

func TestNewRatePlanEmptyName(t *testing.T) {
	mealPlan := MealPlan{Included: false, PricePerPax: 0, Type: 0}
	cancel := CancellationPolicy{IsRefundable: true, Rules: []CancellationRule{}}
	pay := PaymentPolicy{PrepayPercent: 0}
	
	rp, err := NewRatePlan("prop-123", "", mealPlan, cancel, pay)
	
	if err != ErrInvalidRatePlan {
		t.Errorf("Expected ErrInvalidRatePlan, got %v", err)
	}
	if rp != nil {
		t.Error("Expected nil RatePlan on invalid name")
	}
}

func TestNewRatePlanNonRefundableWithRules(t *testing.T) {
	mealPlan := MealPlan{Included: false, PricePerPax: 0, Type: 0}
	cancel := CancellationPolicy{
		IsRefundable: false,
		Rules: []CancellationRule{{HoursBeforeCheckIn: 24, PenaltyType: PenaltyFixedAmount, PenaltyValue: 100}},
	}
	pay := PaymentPolicy{PrepayPercent: 0}
	
	rp, err := NewRatePlan("prop-123", "Non-refundable", mealPlan, cancel, pay)
	
	if err != ErrInvalidRatePlan {
		t.Errorf("Expected ErrInvalidRatePlan for non-refundable with rules, got %v", err)
	}
	if rp != nil {
		t.Error("Expected nil RatePlan on invalid cancellation policy")
	}
}

func TestReconstituteRatePlan(t *testing.T) {
	unitTypeID := "ut-789"
	createdAt := time.Now()
	mealPlan := MealPlan{Included: true, PricePerPax: 0, Type: 1}
	cancel := CancellationPolicy{IsRefundable: true, Rules: []CancellationRule{}}
	pay := PaymentPolicy{PrepayPercent: 30}
	
	rp := ReconstituteRatePlan(
		"rp-456", "prop-123", &unitTypeID,
		"Reconstituted Rate", "A rate plan",
		false, mealPlan, cancel, pay, createdAt,
	)
	
	if rp.ID() != "rp-456" {
		t.Errorf("Expected ID 'rp-456', got '%s'", rp.ID())
	}
	if rp.PropertyID() != "prop-123" {
		t.Errorf("Expected PropertyID 'prop-123', got '%s'", rp.PropertyID())
	}
	if rp.UnitTypeID() == nil || *rp.UnitTypeID() != "ut-789" {
		t.Error("Expected UnitTypeID 'ut-789'")
	}
	if rp.Name() != "Reconstituted Rate" {
		t.Errorf("Expected Name 'Reconstituted Rate', got '%s'", rp.Name())
	}
	if rp.Description() != "A rate plan" {
		t.Errorf("Expected Description 'A rate plan', got '%s'", rp.Description())
	}
	if rp.Active() {
		t.Error("Expected Active to be false")
	}
}

func TestRatePlanGetters(t *testing.T) {
	mealPlan := MealPlan{Included: false, PricePerPax: 1500, Type: 2}
	cancel := CancellationPolicy{
		IsRefundable: true,
		Rules: []CancellationRule{
			{HoursBeforeCheckIn: 48, PenaltyType: PenaltyPercentage, PenaltyValue: 50},
		},
	}
	pay := PaymentPolicy{PrepayPercent: 50}
	
	rp, _ := NewRatePlan("prop-xyz", "Flexible Rate", mealPlan, cancel, pay)
	
	if rp.ID() == "" {
		t.Error("Expected non-empty ID")
	}
	if rp.PropertyID() != "prop-xyz" {
		t.Errorf("Expected PropertyID 'prop-xyz', got '%s'", rp.PropertyID())
	}
	if rp.UnitTypeID() != nil {
		t.Error("Expected UnitTypeID to be nil for new RatePlan")
	}
	if rp.Name() != "Flexible Rate" {
		t.Errorf("Expected Name 'Flexible Rate', got '%s'", rp.Name())
	}
	if rp.Active() != true {
		t.Error("Expected Active to be true")
	}
	if !rp.CancellationPolicy().IsRefundable {
		t.Error("Expected CancellationPolicy.IsRefundable to be true")
	}
	if len(rp.CancellationPolicy().Rules) != 1 {
		t.Errorf("Expected 1 cancellation rule, got %d", len(rp.CancellationPolicy().Rules))
	}
	if rp.PaymentPolicy().PrepayPercent != 50 {
		t.Errorf("Expected PrepayPercent 50, got %d", rp.PaymentPolicy().PrepayPercent)
	}
	if rp.MealPlan().PricePerPax != 1500 {
		t.Errorf("Expected MealPlan.PricePerPax 1500, got %d", rp.MealPlan().PricePerPax)
	}
}

func TestRatePlanUpdate(t *testing.T) {
	rp, _ := NewRatePlan("prop-123", "Original Name", MealPlan{}, CancellationPolicy{IsRefundable: true, Rules: []CancellationRule{}}, PaymentPolicy{})
	
	rp.Update("New Name", "New Description", false)
	
	if rp.Name() != "New Name" {
		t.Errorf("Expected Name 'New Name', got '%s'", rp.Name())
	}
	if rp.Description() != "New Description" {
		t.Errorf("Expected Description 'New Description', got '%s'", rp.Description())
	}
	if rp.Active() {
		t.Error("Expected Active to be false")
	}
}

func TestRatePlanUpdatePartial(t *testing.T) {
	rp, _ := NewRatePlan("prop-123", "Original Name", MealPlan{}, CancellationPolicy{IsRefundable: true, Rules: []CancellationRule{}}, PaymentPolicy{})
	originalName := rp.Name()
	
	rp.Update("", "New Description", false)
	
	if rp.Name() != originalName {
		t.Error("Name should not change when empty string passed")
	}
	if rp.Description() != "New Description" {
		t.Errorf("Expected Description 'New Description', got '%s'", rp.Description())
	}
	if rp.Active() {
		t.Error("Expected Active to be false")
	}
}

func TestCalculateTotalBasic(t *testing.T) {
	rp, _ := NewRatePlan("prop-123", "Standard", MealPlan{Included: false, PricePerPax: 0, Type: 0}, CancellationPolicy{IsRefundable: true, Rules: []CancellationRule{}}, PaymentPolicy{})
	basePrice := vo.NewMoney(10000, "USD")
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)
	
	total := rp.CalculateTotal(basePrice, start, end, 2, 0)
	
	if total.Amount() != 10000 {
		t.Errorf("Expected total 10000, got %d", total.Amount())
	}
	if total.Currency() != "USD" {
		t.Errorf("Expected currency USD, got %s", total.Currency())
	}
}

func TestCalculateTotalWithMealPlanPerPax(t *testing.T) {
	rp, _ := NewRatePlan("prop-123", "With Breakfast", MealPlan{Included: false, PricePerPax: 1500, Type: 1}, CancellationPolicy{IsRefundable: true, Rules: []CancellationRule{}}, PaymentPolicy{})
	basePrice := vo.NewMoney(10000, "EUR")
	start := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 7, 2, 0, 0, 0, 0, time.UTC)
	
	total := rp.CalculateTotal(basePrice, start, end, 2, 1)
	
	if total.Amount() != 14500 {
		t.Errorf("Expected total 14500, got %d", total.Amount())
	}
}

func TestCalculateTotalOneNight(t *testing.T) {
	rp, _ := NewRatePlan("prop-123", "Weekend", MealPlan{Included: true, PricePerPax: 0, Type: 0}, CancellationPolicy{IsRefundable: true, Rules: []CancellationRule{}}, PaymentPolicy{})
	basePrice := vo.NewMoney(5000, "GBP")
	start := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 16, 0, 0, 0, 0, time.UTC)
	
	total := rp.CalculateTotal(basePrice, start, end, 2, 0)
	
	if total.Amount() != 5000 {
		t.Errorf("Expected total 5000, got %d", total.Amount())
	}
}

func TestMealPlanValue(t *testing.T) {
	mp := MealPlan{Included: true, PricePerPax: 2000, Type: 1}
	
	val, err := mp.Value()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	data, ok := val.([]byte)
	if !ok {
		t.Error("Expected []byte value")
	}
	expected := `{"included":true,"price_per_pax":2000,"type":1}`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}

func TestMealPlanScan(t *testing.T) {
	var mp MealPlan
	jsonData := []byte(`{"included":true,"price_per_pax":3000,"type":2}`)
	
	err := mp.Scan(jsonData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !mp.Included {
		t.Error("Expected Included to be true")
	}
	if mp.PricePerPax != 3000 {
		t.Errorf("Expected PricePerPax 3000, got %d", mp.PricePerPax)
	}
	if mp.Type != 2 {
		t.Errorf("Expected Type 2, got %d", mp.Type)
	}
}

func TestMealPlanScanString(t *testing.T) {
	var mp MealPlan
	jsonStr := `{"included":false,"price_per_pax":500,"type":0}`
	
	err := mp.Scan(jsonStr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if mp.Included {
		t.Error("Expected Included to be false")
	}
	if mp.PricePerPax != 500 {
		t.Errorf("Expected PricePerPax 500, got %d", mp.PricePerPax)
	}
}

func TestMealPlanScanNil(t *testing.T) {
	var mp MealPlan
	
	err := mp.Scan(nil)
	if err != nil {
		t.Errorf("Expected no error scanning nil, got %v", err)
	}
	if mp.Included {
		t.Error("Expected Included to be false for nil scan")
	}
}

func TestCancellationPolicyValue(t *testing.T) {
	cp := CancellationPolicy{
		IsRefundable: true,
		Rules: []CancellationRule{
			{HoursBeforeCheckIn: 24, PenaltyType: PenaltyFixedAmount, PenaltyValue: 50},
		},
	}
	
	val, err := cp.Value()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	data, ok := val.([]byte)
	if !ok {
		t.Error("Expected []byte value")
	}
	
	var decoded CancellationPolicy
	err = decoded.Scan(data)
	if err != nil {
		t.Errorf("Failed to decode: %v", err)
	}
	if !decoded.IsRefundable {
		t.Error("Expected IsRefundable to be true")
	}
	if len(decoded.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(decoded.Rules))
	}
}

func TestCancellationPolicyScan(t *testing.T) {
	var cp CancellationPolicy
	jsonData := []byte(`{"is_refundable":false,"rules":[{"hours_before_check_in":48,"penalty_type":"percentage","penalty_value":100}]}`)
	
	err := cp.Scan(jsonData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if cp.IsRefundable {
		t.Error("Expected IsRefundable to be false")
	}
	if len(cp.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(cp.Rules))
	}
	if cp.Rules[0].PenaltyType != PenaltyPercentage {
		t.Errorf("Expected PenaltyType Percentage, got %s", cp.Rules[0].PenaltyType)
	}
}

func TestCancellationPolicyScanNil(t *testing.T) {
	var cp CancellationPolicy
	
	err := cp.Scan(nil)
	if err != nil {
		t.Errorf("Expected no error scanning nil, got %v", err)
	}
	if cp.IsRefundable {
		t.Error("Expected IsRefundable to be false for nil scan")
	}
}

func TestPaymentPolicyValue(t *testing.T) {
	pp := PaymentPolicy{PrepayPercent: 75}
	
	val, err := pp.Value()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	data, ok := val.([]byte)
	if !ok {
		t.Error("Expected []byte value")
	}
	expected := `{"prepay_percent":75}`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}

func TestPaymentPolicyScan(t *testing.T) {
	var pp PaymentPolicy
	jsonData := []byte(`{"prepay_percent":100}`)
	
	err := pp.Scan(jsonData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if pp.PrepayPercent != 100 {
		t.Errorf("Expected PrepayPercent 100, got %d", pp.PrepayPercent)
	}
}

func TestPaymentPolicyScanNil(t *testing.T) {
	var pp PaymentPolicy
	
	err := pp.Scan(nil)
	if err != nil {
		t.Errorf("Expected no error scanning nil, got %v", err)
	}
	if pp.PrepayPercent != 0 {
		t.Errorf("Expected PrepayPercent 0 for nil scan, got %d", pp.PrepayPercent)
	}
}

func TestPenaltyTypeConstants(t *testing.T) {
	if PenaltyFixedAmount != "fixed" {
		t.Errorf("Expected PenaltyFixedAmount 'fixed', got '%s'", PenaltyFixedAmount)
	}
	if PenaltyPercentage != "percentage" {
		t.Errorf("Expected PenaltyPercentage 'percentage', got '%s'", PenaltyPercentage)
	}
	if PenaltyNights != "nights" {
		t.Errorf("Expected PenaltyNights 'nights', got '%s'", PenaltyNights)
	}
}

func TestMealPlanScanDefaultCase(t *testing.T) {
	var mp MealPlan
	err := mp.Scan(12345)
	
	if err == nil {
		t.Error("Expected error for int type")
	}
}

func TestCancellationPolicyScanDefaultCase(t *testing.T) {
	var cp CancellationPolicy
	err := cp.Scan(struct{}{})
	
	if err == nil {
		t.Error("Expected error for struct type")
	}
}

func TestPaymentPolicyScanDefaultCase(t *testing.T) {
	var pp PaymentPolicy
	err := pp.Scan(true)
	
	if err == nil {
		t.Error("Expected error for bool type")
	}
}

func TestCalculateTotalEdgeCases(t *testing.T) {
	rp, _ := NewRatePlan("prop-123", "Test", MealPlan{Included: false, PricePerPax: 0, Type: 0}, CancellationPolicy{IsRefundable: true, Rules: []CancellationRule{}}, PaymentPolicy{})
	basePrice := vo.NewMoney(10000, "USD")
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	
	total := rp.CalculateTotal(basePrice, start, end, 1, 0)
	
	if total.Amount() != 10000 {
		t.Errorf("Expected total 10000, got %d", total.Amount())
	}
}

func TestCalculateTotalMultipleAdultsAndChildren(t *testing.T) {
	rp, _ := NewRatePlan("prop-123", "Family Plan", MealPlan{Included: false, PricePerPax: 2000, Type: 1}, CancellationPolicy{IsRefundable: true, Rules: []CancellationRule{}}, PaymentPolicy{})
	basePrice := vo.NewMoney(10000, "USD")
	start := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 8, 3, 0, 0, 0, 0, time.UTC)
	
	total := rp.CalculateTotal(basePrice, start, end, 2, 2)
	
	if total.Amount() != 26000 {
		t.Errorf("Expected total 26000, got %d", total.Amount())
	}
}

func TestMealPlanScanUnsupportedType(t *testing.T) {
	var mp MealPlan
	err := mp.Scan(map[string]interface{}{"key": "value"})
	
	if err == nil {
		t.Error("Expected error for map type")
	}
}

func TestCancellationPolicyScanUnsupportedType(t *testing.T) {
	var cp CancellationPolicy
	err := cp.Scan([]int{1, 2, 3})
	
	if err == nil {
		t.Error("Expected error for slice type")
	}
}

func TestPaymentPolicyScanUnsupportedType(t *testing.T) {
	var pp PaymentPolicy
	err := pp.Scan(3.14159)
	
	if err == nil {
		t.Error("Expected error for float type")
	}
}

func TestMealPlanScanInvalidJSON(t *testing.T) {
	var mp MealPlan
	err := mp.Scan([]byte(`{invalid json`))
	
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestCancellationPolicyScanInvalidJSON(t *testing.T) {
	var cp CancellationPolicy
	err := cp.Scan([]byte(`{invalid json`))
	
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestPaymentPolicyScanInvalidJSON(t *testing.T) {
	var pp PaymentPolicy
	err := pp.Scan([]byte(`{invalid json`))
	
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestCalculateTotalNoMarkup(t *testing.T) {
	rp, _ := NewRatePlan("prop-123", "Test", MealPlan{Included: false, PricePerPax: 0, Type: 0}, CancellationPolicy{IsRefundable: true, Rules: []CancellationRule{}}, PaymentPolicy{})
	basePrice := vo.NewMoney(25000, "USD")
	start := time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 9, 2, 0, 0, 0, 0, time.UTC)
	
	total := rp.CalculateTotal(basePrice, start, end, 1, 0)
	
	if total.Amount() != 25000 {
		t.Errorf("Expected total 25000, got %d", total.Amount())
	}
}

func TestCancellationPolicyScanChannel(t *testing.T) {
	var cp CancellationPolicy
	ch := make(chan int)
	err := cp.Scan(ch)
	
	if err == nil {
		t.Error("Expected error for channel type")
	}
}

func TestPaymentPolicyScanChannel(t *testing.T) {
	var pp PaymentPolicy
	ch := make(chan string)
	err := pp.Scan(ch)
	
	if err == nil {
		t.Error("Expected error for channel type")
	}
}

func TestCancellationPolicyScanString(t *testing.T) {
	var cp CancellationPolicy
	jsonStr := `{"is_refundable":true,"rules":[{"hours_before_check_in":24,"penalty_type":"fixed","penalty_value":5000}]}`
	
	err := cp.Scan(jsonStr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !cp.IsRefundable {
		t.Error("Expected IsRefundable to be true")
	}
	if len(cp.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(cp.Rules))
	}
}

func TestPaymentPolicyScanString(t *testing.T) {
	var pp PaymentPolicy
	jsonStr := `{"prepay_percent":50}`
	
	err := pp.Scan(jsonStr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if pp.PrepayPercent != 50 {
		t.Errorf("Expected 50, got %d", pp.PrepayPercent)
	}
}
