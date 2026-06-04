package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCreateRatePlanRequest_JSON(t *testing.T) {
	jsonStr := `{
		"property_id":"prop-123",
		"name":"Standard Rate",
		"meal_plan":{"included":true,"price_per_pax":15.00,"type":1},
		"cancellation_policy":{"is_refundable":true,"rules":[]},
		"payment_policy":{"prepay_percent":20}
	}`
	var req CreateRatePlanRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}
	if req.PropertyID != "prop-123" {
		t.Errorf("Expected property_id 'prop-123', got '%s'", req.PropertyID)
	}
	if req.Name != "Standard Rate" {
		t.Errorf("Expected name 'Standard Rate', got '%s'", req.Name)
	}
	if !req.MealPlan.Included {
		t.Error("Expected meal_plan.included to be true")
	}
	if req.MealPlan.PricePerPax != 15.00 {
		t.Errorf("Expected price_per_pax 15.00, got %f", req.MealPlan.PricePerPax)
	}
	if req.MealPlan.Type != 1 {
		t.Errorf("Expected type 1, got %d", req.MealPlan.Type)
	}
	if !req.CancellationPolicy.IsRefundable {
		t.Error("Expected is_refundable to be true")
	}
	if req.PaymentPolicy.PrepayPercent != 20 {
		t.Errorf("Expected prepay_percent 20, got %d", req.PaymentPolicy.PrepayPercent)
	}
}

func TestRatePlanHandler_Create_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/rate-plans", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &RatePlanHandler{service: nil}
	_ = h.Create(c)
	
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestRatePlanHandler_List_MissingPropertyID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/rate-plans", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &RatePlanHandler{service: nil}
	_ = h.List(c)
	
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestRatePlanHandler_Update_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPut, "/rate-plans/rp-123", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("rp-123")

	h := &RatePlanHandler{service: nil}
	_ = h.Update(c)
	
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestUpdateRatePlanRequest_JSON(t *testing.T) {
	jsonStr := `{"name":"Updated Rate","description":"New description","active":false}`
	var req UpdateRatePlanRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}
	if req.Name != "Updated Rate" {
		t.Errorf("Expected name 'Updated Rate', got '%s'", req.Name)
	}
	if req.Description != "New description" {
		t.Errorf("Expected description 'New description', got '%s'", req.Description)
	}
	if req.Active != false {
		t.Error("Expected active to be false")
	}
}

func TestMealPlanDTO_JSON(t *testing.T) {
	jsonStr := `{"included":false,"price_per_pax":25.50,"type":2}`
	var mp MealPlanDTO
	err := json.Unmarshal([]byte(jsonStr), &mp)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}
	if mp.Included != false {
		t.Error("Expected included to be false")
	}
	if mp.PricePerPax != 25.50 {
		t.Errorf("Expected price_per_pax 25.50, got %f", mp.PricePerPax)
	}
	if mp.Type != 2 {
		t.Errorf("Expected type 2, got %d", mp.Type)
	}
}

func TestCreateRatePlanRequest_MinimalJSON(t *testing.T) {
	jsonStr := `{"property_id":"prop-1","name":"Basic"}`
	var req CreateRatePlanRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}
	if req.PropertyID != "prop-1" {
		t.Errorf("Expected property_id 'prop-1', got '%s'", req.PropertyID)
	}
	if req.Name != "Basic" {
		t.Errorf("Expected name 'Basic', got '%s'", req.Name)
	}
}

func TestGetString(t *testing.T) {
	str := "test"
	if getString(&str) != "test" {
		t.Error("Expected 'test'")
	}
	if getString(nil) != "" {
		t.Error("Expected empty string for nil")
	}
}

func BenchmarkCreateRatePlanRequestParsing(b *testing.B) {
	jsonStr := `{"property_id":"prop-123","name":"Standard Rate","meal_plan":{"included":true,"price_per_pax":15.00,"type":1}}`
	for i := 0; i < b.N; i++ {
		var req CreateRatePlanRequest
		_ = json.Unmarshal([]byte(jsonStr), &req)
	}
}
