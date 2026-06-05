package http

import (
	"errors"
	"time"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ecelayes/pms-backend/internal/pricing/application"
	"github.com/ecelayes/pms-backend/internal/pricing/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"github.com/labstack/echo/v4"
)

func newTestPriceRule() *domain.PriceRule {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	dr, _ := vo.NewDateRange(start, end)
	return domain.NewPriceRule("ut-1", dr, vo.NewMoney(10000, "USD"))
}

func TestPricingHandler_BulkUpdate_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid}`
	req := httptest.NewRequest(http.MethodPost, "/pricing", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewPricingHandler(&mockPricingService{})
	_ = h.BulkUpdate(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestPricingHandler_BulkUpdate_InvalidStartDate(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-1","start":"bad","end":"2025-01-31","price":100,"currency":"USD"}`
	req := httptest.NewRequest(http.MethodPost, "/pricing", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewPricingHandler(&mockPricingService{})
	_ = h.BulkUpdate(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestPricingHandler_BulkUpdate_InvalidEndDate(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-1","start":"2025-01-01","end":"bad","price":100,"currency":"USD"}`
	req := httptest.NewRequest(http.MethodPost, "/pricing", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewPricingHandler(&mockPricingService{})
	_ = h.BulkUpdate(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestPricingHandler_BulkUpdate_NegativePrice(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-1","start":"2025-01-01","end":"2025-01-31","price":-10,"currency":"USD"}`
	req := httptest.NewRequest(http.MethodPost, "/pricing", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewPricingHandler(&mockPricingService{})
	_ = h.BulkUpdate(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestPricingHandler_BulkUpdate_DefaultCurrency(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-1","start":"2025-01-01","end":"2025-01-31","price":100}`
	req := httptest.NewRequest(http.MethodPost, "/pricing", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPricingService{}
	h := NewPricingHandler(svc)
	_ = h.BulkUpdate(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestPricingHandler_BulkUpdate_InvalidDateRange(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-1","start":"2025-01-01","end":"2025-01-31","price":100,"currency":"USD"}`
	req := httptest.NewRequest(http.MethodPost, "/pricing", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPricingService{setRuleErr: vo.ErrinvalidDateRange}
	h := NewPricingHandler(svc)
	_ = h.BulkUpdate(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestPricingHandler_BulkUpdate_Error(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-1","start":"2025-01-01","end":"2025-01-31","price":100,"currency":"USD"}`
	req := httptest.NewRequest(http.MethodPost, "/pricing", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPricingService{setRuleErr: errors.New("db error")}
	h := NewPricingHandler(svc)
	_ = h.BulkUpdate(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestPricingHandler_GetRules_MissingParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/pricing", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewPricingHandler(&mockPricingService{})
	_ = h.GetRules(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestPricingHandler_GetRules_NilResult(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/pricing?unit_type_id=ut-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPricingService{getRulesResult: nil}
	h := NewPricingHandler(svc)
	_ = h.GetRules(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestPricingHandler_GetRules_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/pricing?unit_type_id=ut-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPricingService{getRulesResult: []*domain.PriceRule{newTestPriceRule()}}
	h := NewPricingHandler(svc)
	_ = h.GetRules(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestPricingHandler_GetRules_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/pricing?unit_type_id=ut-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPricingService{getRulesErr: errors.New("db error")}
	h := NewPricingHandler(svc)
	_ = h.GetRules(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestPricingHandler_DeleteRule_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/pricing/pr-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("pr-1")

	svc := &mockPricingService{}
	h := NewPricingHandler(svc)
	_ = h.DeleteRule(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestPricingHandler_DeleteRule_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/pricing/pr-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("pr-1")

	svc := &mockPricingService{deleteRuleErr: application.ErrNotFound}
	h := NewPricingHandler(svc)
	_ = h.DeleteRule(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestPricingHandler_DeleteRule_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/pricing/pr-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("pr-1")

	svc := &mockPricingService{deleteRuleErr: errors.New("db error")}
	h := NewPricingHandler(svc)
	_ = h.DeleteRule(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestPricingHandler_NewPricingHandler(t *testing.T) {
	svc := &mockPricingService{}
	h := NewPricingHandler(svc)
	if h == nil || h.service == nil {
		t.Fatal("handler or service nil")
	}
}
