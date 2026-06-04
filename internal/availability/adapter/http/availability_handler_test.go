package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestAvailabilityHandler_Get_MissingStart(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/availability?end=2024-06-03", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AvailabilityHandler{service: nil}
	_ = h.Get(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_MissingEnd(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/availability?start=2024-06-01", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AvailabilityHandler{service: nil}
	_ = h.Get(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_InvalidStartDate(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/availability?start=invalid&end=2024-06-03", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AvailabilityHandler{service: nil}
	_ = h.Get(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_InvalidEndDate(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/availability?start=2024-06-01&end=invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AvailabilityHandler{service: nil}
	_ = h.Get(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestAvailabilityHandler_Get_StartAfterEnd(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/availability?start=2024-06-10&end=2024-06-03", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AvailabilityHandler{service: nil}
	_ = h.Get(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
