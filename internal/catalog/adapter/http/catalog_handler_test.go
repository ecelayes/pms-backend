package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ecelayes/pms-backend/internal/catalog/application"
	"github.com/ecelayes/pms-backend/internal/catalog/domain"
	"github.com/labstack/echo/v4"
)

func TestCatalogHandler_CreateAmenity_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"WiFi","description":"Free WiFi","icon":"wifi"}`
	req := httptest.NewRequest(http.MethodPost, "/amenities", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockCatalogService{amenityCreateResult: "a-1"}
	h := NewCatalogHandler(svc)
	_ = h.CreateAmenity(c)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d", rec.Code)
	}
}

func TestCatalogHandler_CreateAmenity_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid}`
	req := httptest.NewRequest(http.MethodPost, "/amenities", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewCatalogHandler(&mockCatalogService{})
	_ = h.CreateAmenity(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestCatalogHandler_CreateAmenity_Duplicate(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"WiFi"}`
	req := httptest.NewRequest(http.MethodPost, "/amenities", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockCatalogService{amenityCreateErr: errors.New("duplicate key")}
	h := NewCatalogHandler(svc)
	_ = h.CreateAmenity(c)

	if rec.Code != http.StatusConflict {
		t.Errorf("Expected 409, got %d", rec.Code)
	}
}

func TestCatalogHandler_CreateAmenity_Invalid(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"WiFi"}`
	req := httptest.NewRequest(http.MethodPost, "/amenities", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockCatalogService{amenityCreateErr: errors.New("invalid input")}
	h := NewCatalogHandler(svc)
	_ = h.CreateAmenity(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestCatalogHandler_CreateAmenity_Internal(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"WiFi"}`
	req := httptest.NewRequest(http.MethodPost, "/amenities", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockCatalogService{amenityCreateErr: errors.New("db error")}
	h := NewCatalogHandler(svc)
	_ = h.CreateAmenity(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestCatalogHandler_GetAllAmenities_Default(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/amenities", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockCatalogService{amenityListResult: nil, amenityListTotal: 0}
	h := NewCatalogHandler(svc)
	_ = h.GetAllAmenities(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestCatalogHandler_GetAllAmenities_WithParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/amenities?page=2&limit=20", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockCatalogService{amenityListResult: nil, amenityListTotal: 0}
	h := NewCatalogHandler(svc)
	_ = h.GetAllAmenities(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestCatalogHandler_GetAllAmenities_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/amenities", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockCatalogService{amenityListErr: errors.New("db error")}
	h := NewCatalogHandler(svc)
	_ = h.GetAllAmenities(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestCatalogHandler_GetAmenityByID_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/amenities/a-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("a-1")

	svc := &mockCatalogService{amenityGetErr: application.ErrNotFound}
	h := NewCatalogHandler(svc)
	_ = h.GetAmenityByID(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestCatalogHandler_GetAmenityByID_Internal(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/amenities/a-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("a-1")

	svc := &mockCatalogService{amenityGetErr: errors.New("db error")}
	h := NewCatalogHandler(svc)
	_ = h.GetAmenityByID(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestCatalogHandler_UpdateAmenity_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated","description":"d","icon":"i"}`
	req := httptest.NewRequest(http.MethodPut, "/amenities/a-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("a-1")

	svc := &mockCatalogService{}
	h := NewCatalogHandler(svc)
	_ = h.UpdateAmenity(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestCatalogHandler_UpdateAmenity_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid}`
	req := httptest.NewRequest(http.MethodPut, "/amenities/a-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewCatalogHandler(&mockCatalogService{})
	_ = h.UpdateAmenity(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestCatalogHandler_UpdateAmenity_NotFound(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/amenities/a-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("a-1")

	svc := &mockCatalogService{amenityUpdateErr: application.ErrNotFound}
	h := NewCatalogHandler(svc)
	_ = h.UpdateAmenity(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestCatalogHandler_UpdateAmenity_Internal(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/amenities/a-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("a-1")

	svc := &mockCatalogService{amenityUpdateErr: errors.New("db error")}
	h := NewCatalogHandler(svc)
	_ = h.UpdateAmenity(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestCatalogHandler_DeleteAmenity_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/amenities/a-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("a-1")

	svc := &mockCatalogService{}
	h := NewCatalogHandler(svc)
	_ = h.DeleteAmenity(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestCatalogHandler_DeleteAmenity_Internal(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/amenities/a-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("a-1")

	svc := &mockCatalogService{amenityDeleteErr: errors.New("db error")}
	h := NewCatalogHandler(svc)
	_ = h.DeleteAmenity(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestCatalogHandler_CreateService_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Spa","description":"Massage","icon":"spa"}`
	req := httptest.NewRequest(http.MethodPost, "/services", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockCatalogService{guestCreateResult: "s-1"}
	h := NewCatalogHandler(svc)
	_ = h.CreateService(c)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d", rec.Code)
	}
}

func TestCatalogHandler_CreateService_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid}`
	req := httptest.NewRequest(http.MethodPost, "/services", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewCatalogHandler(&mockCatalogService{})
	_ = h.CreateService(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestCatalogHandler_CreateService_Duplicate(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Spa"}`
	req := httptest.NewRequest(http.MethodPost, "/services", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockCatalogService{guestCreateErr: errors.New("duplicate key")}
	h := NewCatalogHandler(svc)
	_ = h.CreateService(c)

	if rec.Code != http.StatusConflict {
		t.Errorf("Expected 409, got %d", rec.Code)
	}
}

func TestCatalogHandler_CreateService_Invalid(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Spa"}`
	req := httptest.NewRequest(http.MethodPost, "/services", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockCatalogService{guestCreateErr: errors.New("invalid input")}
	h := NewCatalogHandler(svc)
	_ = h.CreateService(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestCatalogHandler_CreateService_Internal(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Spa"}`
	req := httptest.NewRequest(http.MethodPost, "/services", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockCatalogService{guestCreateErr: errors.New("db error")}
	h := NewCatalogHandler(svc)
	_ = h.CreateService(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestCatalogHandler_GetAllServices_Default(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockCatalogService{guestListResult: nil, guestListTotal: 0}
	h := NewCatalogHandler(svc)
	_ = h.GetAllServices(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestCatalogHandler_GetAllServices_WithParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/services?page=1&limit=20", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockCatalogService{guestListResult: nil, guestListTotal: 0}
	h := NewCatalogHandler(svc)
	_ = h.GetAllServices(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestCatalogHandler_GetAllServices_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockCatalogService{guestListErr: errors.New("db error")}
	h := NewCatalogHandler(svc)
	_ = h.GetAllServices(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestCatalogHandler_GetServiceByID_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/services/s-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("s-1")

	svc := &mockCatalogService{guestGetErr: application.ErrNotFound}
	h := NewCatalogHandler(svc)
	_ = h.GetServiceByID(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestCatalogHandler_GetServiceByID_Internal(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/services/s-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("s-1")

	svc := &mockCatalogService{guestGetErr: errors.New("db error")}
	h := NewCatalogHandler(svc)
	_ = h.GetServiceByID(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestCatalogHandler_UpdateService_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated","description":"d","icon":"i"}`
	req := httptest.NewRequest(http.MethodPut, "/services/s-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("s-1")

	svc := &mockCatalogService{}
	h := NewCatalogHandler(svc)
	_ = h.UpdateService(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestCatalogHandler_UpdateService_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid}`
	req := httptest.NewRequest(http.MethodPut, "/services/s-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewCatalogHandler(&mockCatalogService{})
	_ = h.UpdateService(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestCatalogHandler_UpdateService_NotFound(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/services/s-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("s-1")

	svc := &mockCatalogService{guestUpdateErr: application.ErrNotFound}
	h := NewCatalogHandler(svc)
	_ = h.UpdateService(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestCatalogHandler_UpdateService_Internal(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/services/s-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("s-1")

	svc := &mockCatalogService{guestUpdateErr: errors.New("db error")}
	h := NewCatalogHandler(svc)
	_ = h.UpdateService(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestCatalogHandler_DeleteService_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/services/s-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("s-1")

	svc := &mockCatalogService{}
	h := NewCatalogHandler(svc)
	_ = h.DeleteService(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestCatalogHandler_DeleteService_Internal(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/services/s-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("s-1")

	svc := &mockCatalogService{guestDeleteErr: errors.New("db error")}
	h := NewCatalogHandler(svc)
	_ = h.DeleteService(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestCatalogHandler_NewCatalogHandler(t *testing.T) {
	svc := &mockCatalogService{}
	h := NewCatalogHandler(svc)
	if h == nil || h.service == nil {
		t.Fatal("handler or service nil")
	}
}

// --- toAmenityDTO / toGuestServiceDTO / GetByID success paths ---

func TestCatalogHandler_toAmenityDTO(t *testing.T) {
	a, _ := domain.NewAmenity("WiFi", "Fast internet", "wifi-icon")
	h := NewCatalogHandler(&mockCatalogService{})
	d := h.toAmenityDTO(a)
	if d.ID != a.ID() || d.Name != "WiFi" || d.Description != "Fast internet" || d.Icon != "wifi-icon" {
		t.Errorf("DTO mismatch: %+v", d)
	}
}

func TestCatalogHandler_toGuestServiceDTO(t *testing.T) {
	s, _ := domain.NewGuestService("Breakfast", "Hot", "egg-icon")
	h := NewCatalogHandler(&mockCatalogService{})
	d := h.toGuestServiceDTO(s)
	if d.ID != s.ID() || d.Name != "Breakfast" || d.Description != "Hot" || d.Icon != "egg-icon" {
		t.Errorf("DTO mismatch: %+v", d)
	}
}

func TestCatalogHandler_GetAmenityByID_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/amenities/a-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("a-1")

	a, _ := domain.NewAmenity("WiFi", "d", "i")
	svc := &mockCatalogService{amenityGetResult: a}
	h := NewCatalogHandler(svc)
	_ = h.GetAmenityByID(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestCatalogHandler_GetServiceByID_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/services/s-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("s-1")

	s, _ := domain.NewGuestService("Spa", "d", "i")
	svc := &mockCatalogService{guestGetResult: s}
	h := NewCatalogHandler(svc)
	_ = h.GetServiceByID(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

// --- GetAll with non-empty list (exercises toAmenityDTO/toGuestServiceDTO loop) ---

func TestCatalogHandler_GetAllAmenities_WithData(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/amenities", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	a, _ := domain.NewAmenity("WiFi", "desc", "icon")
	svc := &mockCatalogService{amenityListResult: []*domain.Amenity{a}, amenityListTotal: 1}
	h := NewCatalogHandler(svc)
	_ = h.GetAllAmenities(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestCatalogHandler_GetAllServices_WithData(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	s, _ := domain.NewGuestService("Spa", "desc", "icon")
	svc := &mockCatalogService{guestListResult: []*domain.GuestService{s}, guestListTotal: 1}
	h := NewCatalogHandler(svc)
	_ = h.GetAllServices(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}
