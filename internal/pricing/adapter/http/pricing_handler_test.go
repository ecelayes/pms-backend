package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestSetPriceRequest_JSON(t *testing.T) {
	jsonStr := `{"unit_type_id":"ut-123","start":"2024-06-01","end":"2024-06-05","price":150.50,"currency":"EUR"}`
	var req SetPriceRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}
	if req.UnitTypeID != "ut-123" {
		t.Errorf("Expected unit_type_id 'ut-123', got '%s'", req.UnitTypeID)
	}
	if req.Start != "2024-06-01" {
		t.Errorf("Expected start '2024-06-01', got '%s'", req.Start)
	}
	if req.End != "2024-06-05" {
		t.Errorf("Expected end '2024-06-05', got '%s'", req.End)
	}
	if req.Price != 150.50 {
		t.Errorf("Expected price 150.50, got %f", req.Price)
	}
	if req.Currency != "EUR" {
		t.Errorf("Expected currency 'EUR', got '%s'", req.Currency)
	}
}

func TestBulkUpdate_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/pricing/bulk", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &PricingHandler{service: nil}
	_ = h.BulkUpdate(c)
	
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestBulkUpdate_InvalidStartDate(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-123","start":"invalid-date","end":"2024-06-05","price":150.00,"currency":"USD"}`
	req := httptest.NewRequest(http.MethodPost, "/pricing/bulk", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &PricingHandler{service: nil}
	_ = h.BulkUpdate(c)
	
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestBulkUpdate_InvalidEndDate(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-123","start":"2024-06-01","end":"invalid-date","price":150.00,"currency":"USD"}`
	req := httptest.NewRequest(http.MethodPost, "/pricing/bulk", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &PricingHandler{service: nil}
	_ = h.BulkUpdate(c)
	
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestBulkUpdate_NegativePrice(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-123","start":"2024-06-01","end":"2024-06-05","price":-50.00,"currency":"USD"}`
	req := httptest.NewRequest(http.MethodPost, "/pricing/bulk", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &PricingHandler{service: nil}
	_ = h.BulkUpdate(c)
	
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestGetRules_MissingParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/pricing/rules", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &PricingHandler{service: nil}
	_ = h.GetRules(c)
	
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestBulkUpdate_ValidJSONParsing(t *testing.T) {
	e := echo.New()
	// Only test parsing - use negative price to stop before service call
	reqBody := `{"unit_type_id":"ut-123","start":"2024-06-01","end":"2024-06-05","price":-1,"currency":"USD"}`
	req := httptest.NewRequest(http.MethodPost, "/pricing/bulk", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &PricingHandler{service: nil}
	_ = h.BulkUpdate(c)
	
	// Should fail on negative price validation before reaching service
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func BenchmarkSetPriceRequestParsing(b *testing.B) {
	jsonStr := `{"unit_type_id":"ut-123","start":"2024-06-01","end":"2024-06-05","price":150.50,"currency":"EUR"}`
	for i := 0; i < b.N; i++ {
		var req SetPriceRequest
		_ = json.Unmarshal([]byte(jsonStr), &req)
	}
}
