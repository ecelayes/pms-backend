package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ecelayes/pms-backend/internal/shared/dto"
	"github.com/labstack/echo/v4"
)

func TestAvailabilityHandler_Get_MissingDates(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/availability", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewAvailabilityHandler(&mockAvailabilityService{})
	_ = h.Get(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_InvalidStartDate(t *testing.T) {
	e := echo.New()
	q := url.Values{}
	q.Set("start", "bad")
	q.Set("end", "2025-12-31")
	req := httptest.NewRequest(http.MethodGet, "/availability?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewAvailabilityHandler(&mockAvailabilityService{})
	_ = h.Get(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_InvalidEndDate(t *testing.T) {
	e := echo.New()
	q := url.Values{}
	q.Set("start", "2025-01-01")
	q.Set("end", "bad")
	req := httptest.NewRequest(http.MethodGet, "/availability?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewAvailabilityHandler(&mockAvailabilityService{})
	_ = h.Get(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_StartAfterEnd(t *testing.T) {
	e := echo.New()
	q := url.Values{}
	q.Set("start", "2025-12-31")
	q.Set("end", "2025-01-01")
	req := httptest.NewRequest(http.MethodGet, "/availability?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewAvailabilityHandler(&mockAvailabilityService{})
	_ = h.Get(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_StartEqualsEnd(t *testing.T) {
	e := echo.New()
	q := url.Values{}
	q.Set("start", "2025-01-01")
	q.Set("end", "2025-01-01")
	req := httptest.NewRequest(http.MethodGet, "/availability?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewAvailabilityHandler(&mockAvailabilityService{})
	_ = h.Get(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_EmptyResults(t *testing.T) {
	e := echo.New()
	q := url.Values{}
	q.Set("start", "2025-06-01")
	q.Set("end", "2025-06-10")
	req := httptest.NewRequest(http.MethodGet, "/availability?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockAvailabilityService{searchResult: nil}
	h := NewAvailabilityHandler(svc)
	_ = h.Get(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_WithPagination(t *testing.T) {
	e := echo.New()
	q := url.Values{}
	q.Set("start", "2025-06-01")
	q.Set("end", "2025-06-10")
	q.Set("page", "1")
	q.Set("limit", "5")
	req := httptest.NewRequest(http.MethodGet, "/availability?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockAvailabilityService{searchResult: nil}
	h := NewAvailabilityHandler(svc)
	_ = h.Get(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_PageBeyondResults(t *testing.T) {
	e := echo.New()
	q := url.Values{}
	q.Set("start", "2025-06-01")
	q.Set("end", "2025-06-10")
	q.Set("page", "100")
	q.Set("limit", "10")
	req := httptest.NewRequest(http.MethodGet, "/availability?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockAvailabilityService{searchResult: nil}
	h := NewAvailabilityHandler(svc)
	_ = h.Get(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_WithRooms(t *testing.T) {
	e := echo.New()
	q := url.Values{}
	q.Set("start", "2025-06-01")
	q.Set("end", "2025-06-10")
	q.Set("rooms", "2")
	q.Set("adults", "3")
	q.Set("children", "1")
	req := httptest.NewRequest(http.MethodGet, "/availability?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockAvailabilityService{searchResult: nil}
	h := NewAvailabilityHandler(svc)
	_ = h.Get(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_WithResults(t *testing.T) {
	e := echo.New()
	q := url.Values{}
	q.Set("start", "2025-06-01")
	q.Set("end", "2025-06-10")
	req := httptest.NewRequest(http.MethodGet, "/availability?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockAvailabilityService{searchResult: []dto.AvailabilityResult{
		{UnitTypeID: "ut-1", AvailableQty: 5, TotalPrice: 100.0, Currency: "USD"},
	}}
	h := NewAvailabilityHandler(svc)
	_ = h.Get(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_InvalidPageAndLimit(t *testing.T) {
	e := echo.New()
	q := url.Values{}
	q.Set("start", "2025-06-01")
	q.Set("end", "2025-06-10")
	q.Set("page", "abc")
	q.Set("limit", "xyz")
	req := httptest.NewRequest(http.MethodGet, "/availability?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockAvailabilityService{searchResult: nil}
	h := NewAvailabilityHandler(svc)
	_ = h.Get(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_SearchError(t *testing.T) {
	e := echo.New()
	q := url.Values{}
	q.Set("start", "2025-06-01")
	q.Set("end", "2025-06-10")
	req := httptest.NewRequest(http.MethodGet, "/availability?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockAvailabilityService{searchErr: errors.New("db error")}
	h := NewAvailabilityHandler(svc)
	_ = h.Get(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_PageClamping(t *testing.T) {
	e := echo.New()
	q := url.Values{}
	q.Set("start", "2025-06-01")
	q.Set("end", "2025-06-10")
	q.Set("page", "2")
	q.Set("limit", "3")
	req := httptest.NewRequest(http.MethodGet, "/availability?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockAvailabilityService{searchResult: []dto.AvailabilityResult{
		{UnitTypeID: "ut-1"}, {UnitTypeID: "ut-2"}, {UnitTypeID: "ut-3"}, {UnitTypeID: "ut-4"}, {UnitTypeID: "ut-5"},
	}}
	h := NewAvailabilityHandler(svc)
	_ = h.Get(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_PageExactlyAtEnd(t *testing.T) {
	e := echo.New()
	q := url.Values{}
	q.Set("start", "2025-06-01")
	q.Set("end", "2025-06-10")
	q.Set("page", "2")
	q.Set("limit", "3")
	req := httptest.NewRequest(http.MethodGet, "/availability?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockAvailabilityService{searchResult: []dto.AvailabilityResult{
		{UnitTypeID: "ut-1"}, {UnitTypeID: "ut-2"}, {UnitTypeID: "ut-3"},
		{UnitTypeID: "ut-4"}, {UnitTypeID: "ut-5"}, {UnitTypeID: "ut-6"},
	}}
	h := NewAvailabilityHandler(svc)
	_ = h.Get(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_NewAvailabilityHandler(t *testing.T) {
	svc := &mockAvailabilityService{}
	h := NewAvailabilityHandler(svc)
	if h == nil || h.service == nil {
		t.Fatal("handler or service nil")
	}
}
