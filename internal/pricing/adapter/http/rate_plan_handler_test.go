package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ecelayes/pms-backend/internal/pricing/application"
	"github.com/ecelayes/pms-backend/internal/pricing/domain"
	"github.com/labstack/echo/v4"
)

func newTestRatePlan() *domain.RatePlan {
	mp := domain.MealPlan{Included: true, PricePerPax: 1000, Type: 1}
	cp := domain.CancellationPolicy{}
	pp := domain.PaymentPolicy{}
	r, _ := domain.NewRatePlan("p-1", "Standard", mp, cp, pp)
	return r
}

func TestRatePlanHandler_Create_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid}`
	req := httptest.NewRequest(http.MethodPost, "/rate-plans", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewRatePlanHandler(&mockPricingService{})
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestRatePlanHandler_Create_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"property_id":"p-1","name":"Standard","meal_plan":{"included":true,"price_per_pax":10.0,"type":1},"cancellation_policy":{},"payment_policy":{}}`
	req := httptest.NewRequest(http.MethodPost, "/rate-plans", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPricingService{createRPResult: "rp-1"}
	h := NewRatePlanHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d", rec.Code)
	}
}

func TestRatePlanHandler_Create_Invalid(t *testing.T) {
	e := echo.New()
	reqBody := `{"property_id":"p-1","name":"Standard","meal_plan":{},"cancellation_policy":{},"payment_policy":{}}`
	req := httptest.NewRequest(http.MethodPost, "/rate-plans", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPricingService{createRPErr: errors.New("invalid name")}
	h := NewRatePlanHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestRatePlanHandler_Create_Internal(t *testing.T) {
	e := echo.New()
	reqBody := `{"property_id":"p-1","name":"Standard","meal_plan":{},"cancellation_policy":{},"payment_policy":{}}`
	req := httptest.NewRequest(http.MethodPost, "/rate-plans", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPricingService{createRPErr: errors.New("db error")}
	h := NewRatePlanHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestRatePlanHandler_List_MissingProperty(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/rate-plans", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewRatePlanHandler(&mockPricingService{})
	_ = h.List(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestRatePlanHandler_List_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/rate-plans?property_id=p-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPricingService{listRPResult: []*domain.RatePlan{newTestRatePlan()}}
	h := NewRatePlanHandler(svc)
	_ = h.List(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestRatePlanHandler_List_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/rate-plans?property_id=p-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPricingService{listRPErr: errors.New("db error")}
	h := NewRatePlanHandler(svc)
	_ = h.List(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestRatePlanHandler_GetByID_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/rate-plans/rp-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("rp-1")

	svc := &mockPricingService{getRPResult: newTestRatePlan()}
	h := NewRatePlanHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestRatePlanHandler_GetByID_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/rate-plans/rp-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("rp-1")

	svc := &mockPricingService{getRPErr: application.ErrNotFound}
	h := NewRatePlanHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestRatePlanHandler_GetByID_Internal(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/rate-plans/rp-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("rp-1")

	svc := &mockPricingService{getRPErr: errors.New("db error")}
	h := NewRatePlanHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestRatePlanHandler_Update_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated","description":"d","active":true}`
	req := httptest.NewRequest(http.MethodPut, "/rate-plans/rp-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("rp-1")

	svc := &mockPricingService{}
	h := NewRatePlanHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestRatePlanHandler_Update_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid}`
	req := httptest.NewRequest(http.MethodPut, "/rate-plans/rp-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewRatePlanHandler(&mockPricingService{})
	_ = h.Update(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestRatePlanHandler_Update_NotFound(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated","description":"d","active":true}`
	req := httptest.NewRequest(http.MethodPut, "/rate-plans/rp-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("rp-1")

	svc := &mockPricingService{updateRPErr: application.ErrNotFound}
	h := NewRatePlanHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestRatePlanHandler_Update_Internal(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated","description":"d","active":true}`
	req := httptest.NewRequest(http.MethodPut, "/rate-plans/rp-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("rp-1")

	svc := &mockPricingService{updateRPErr: errors.New("db error")}
	h := NewRatePlanHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestRatePlanHandler_Delete_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/rate-plans/rp-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("rp-1")

	svc := &mockPricingService{}
	h := NewRatePlanHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestRatePlanHandler_Delete_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/rate-plans/rp-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("rp-1")

	svc := &mockPricingService{deleteRPErr: application.ErrNotFound}
	h := NewRatePlanHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestRatePlanHandler_Delete_Internal(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/rate-plans/rp-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("rp-1")

	svc := &mockPricingService{deleteRPErr: errors.New("db error")}
	h := NewRatePlanHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestRatePlanHandler_NewRatePlanHandler(t *testing.T) {
	svc := &mockPricingService{}
	h := NewRatePlanHandler(svc)
	if h == nil || h.service == nil {
		t.Fatal("handler or service nil")
	}
}

func TestRatePlanHandler_GetByID_Success_WithUnitType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/rate-plans/rp-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("rp-1")

	utID := "ut-1"
	rp := domain.ReconstituteRatePlan(
		"rp-1", "p-1", &utID,
		"Standard", "desc", true,
		domain.MealPlan{Included: true, PricePerPax: 1000, Type: 1},
		domain.CancellationPolicy{},
		domain.PaymentPolicy{},
		time.Now(),
	)
	svc := &mockPricingService{getRPResult: rp}
	h := NewRatePlanHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}
