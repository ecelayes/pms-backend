package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ecelayes/pms-backend/internal/catalog/application"
	"github.com/ecelayes/pms-backend/internal/catalog/domain"
	vo "github.com/ecelayes/pms-backend/internal/shared/vo"
	"github.com/labstack/echo/v4"
)

func newTestProperty() *domain.Property {
	p, _ := domain.NewProperty("org-1", "Hotel A", "HA", domain.PropertyTypeHotel)
	return p
}

func TestPropertyHandler_Create_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"organization_id":"org-1","name":"Hotel A","code":"HA","type":"hotel"}`
	req := httptest.NewRequest(http.MethodPost, "/properties", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPropertyService{createResult: "p-1"}
	h := NewPropertyHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d", rec.Code)
	}
}

func TestPropertyHandler_Create_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/properties", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewPropertyHandler(&mockPropertyService{})
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestPropertyHandler_Create_Duplicate(t *testing.T) {
	e := echo.New()
	reqBody := `{"organization_id":"org-1","name":"Hotel A","code":"HA","type":"hotel"}`
	req := httptest.NewRequest(http.MethodPost, "/properties", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPropertyService{createErr: errors.New("duplicate key violates unique constraint")}
	h := NewPropertyHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusConflict {
		t.Errorf("Expected 409, got %d", rec.Code)
	}
}

func TestPropertyHandler_Create_Invalid(t *testing.T) {
	e := echo.New()
	reqBody := `{"organization_id":"org-1","name":"Hotel A","code":"HA","type":"hotel"}`
	req := httptest.NewRequest(http.MethodPost, "/properties", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPropertyService{createErr: errors.New("invalid input: name required")}
	h := NewPropertyHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestPropertyHandler_Create_Internal(t *testing.T) {
	e := echo.New()
	reqBody := `{"organization_id":"org-1","name":"Hotel A","code":"HA","type":"hotel"}`
	req := httptest.NewRequest(http.MethodPost, "/properties", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPropertyService{createErr: errors.New("db error")}
	h := NewPropertyHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestPropertyHandler_GetAll_Default(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/properties", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPropertyService{listResult: []*domain.Property{newTestProperty()}, listTotal: 1}
	h := NewPropertyHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestPropertyHandler_GetAll_WithParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/properties?page=2&limit=20", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPropertyService{listResult: []*domain.Property{}, listTotal: 0}
	h := NewPropertyHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestPropertyHandler_GetAll_Empty(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/properties", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPropertyService{listResult: nil, listTotal: 0}
	h := NewPropertyHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestPropertyHandler_GetAll_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/properties", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockPropertyService{listErr: errors.New("db error")}
	h := NewPropertyHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestPropertyHandler_GetByID_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/properties/p-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("p-1")

	svc := &mockPropertyService{getResult: newTestProperty()}
	h := NewPropertyHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestPropertyHandler_GetByID_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/properties/p-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("p-1")

	svc := &mockPropertyService{getErr: application.ErrNotFound}
	h := NewPropertyHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestPropertyHandler_GetByID_Internal(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/properties/p-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("p-1")

	svc := &mockPropertyService{getErr: errors.New("db error")}
	h := NewPropertyHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestPropertyHandler_Update_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated","code":"UPD","type":"hotel"}`
	req := httptest.NewRequest(http.MethodPut, "/properties/p-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("p-1")

	svc := &mockPropertyService{}
	h := NewPropertyHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestPropertyHandler_Update_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPut, "/properties/123", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewPropertyHandler(&mockPropertyService{})
	_ = h.Update(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestPropertyHandler_Update_NotFound(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated","code":"UPD","type":"hotel"}`
	req := httptest.NewRequest(http.MethodPut, "/properties/p-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("p-1")

	svc := &mockPropertyService{updateErr: application.ErrNotFound}
	h := NewPropertyHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestPropertyHandler_Update_Internal(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated","code":"UPD","type":"hotel"}`
	req := httptest.NewRequest(http.MethodPut, "/properties/p-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("p-1")

	svc := &mockPropertyService{updateErr: errors.New("db error")}
	h := NewPropertyHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestPropertyHandler_Delete_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/properties/p-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("p-1")

	svc := &mockPropertyService{}
	h := NewPropertyHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestPropertyHandler_Delete_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/properties/p-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("p-1")

	svc := &mockPropertyService{deleteErr: application.ErrNotFound}
	h := NewPropertyHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestPropertyHandler_Delete_Internal(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/properties/p-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("p-1")

	svc := &mockPropertyService{deleteErr: errors.New("db error")}
	h := NewPropertyHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestPropertyHandler_NewPropertyHandler(t *testing.T) {
	svc := &mockPropertyService{}
	h := NewPropertyHandler(svc)
	if h == nil || h.service == nil {
		t.Fatal("handler or service nil")
	}
}

// silence import warning
var _ = vo.Money{}
