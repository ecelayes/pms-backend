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

func newTestUnit() *domain.Unit {
	u := domain.NewUnit("p-1", "ut-1", "Room 101")
	return u
}

func TestUnitHandler_Create_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"property_id":"p-1","unit_type_id":"ut-1","name":"Room 101"}`
	req := httptest.NewRequest(http.MethodPost, "/units", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockUnitService{createResult: "u-1"}
	h := NewUnitHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d", rec.Code)
	}
}

func TestUnitHandler_Create_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid}`
	req := httptest.NewRequest(http.MethodPost, "/units", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewUnitHandler(&mockUnitService{})
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestUnitHandler_Create_Duplicate(t *testing.T) {
	e := echo.New()
	reqBody := `{"property_id":"p-1","unit_type_id":"ut-1","name":"Room 101"}`
	req := httptest.NewRequest(http.MethodPost, "/units", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockUnitService{createErr: errors.New("duplicate key")}
	h := NewUnitHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusConflict {
		t.Errorf("Expected 409, got %d", rec.Code)
	}
}

func TestUnitHandler_Create_Internal(t *testing.T) {
	e := echo.New()
	reqBody := `{"property_id":"p-1","unit_type_id":"ut-1","name":"Room 101"}`
	req := httptest.NewRequest(http.MethodPost, "/units", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockUnitService{createErr: errors.New("db error")}
	h := NewUnitHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestUnitHandler_GetAll_MissingProperty(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/units", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewUnitHandler(&mockUnitService{})
	_ = h.GetAll(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestUnitHandler_GetAll_Default(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/units?property_id=p-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockUnitService{listResult: []*domain.Unit{newTestUnit()}, listTotal: 1}
	h := NewUnitHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestUnitHandler_GetAll_WithParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/units?property_id=p-1&page=2&limit=5", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockUnitService{listResult: []*domain.Unit{}, listTotal: 0}
	h := NewUnitHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestUnitHandler_GetAll_Empty(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/units?property_id=p-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockUnitService{listResult: nil, listTotal: 0}
	h := NewUnitHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestUnitHandler_GetAll_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/units?property_id=p-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockUnitService{listErr: errors.New("db error")}
	h := NewUnitHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestUnitHandler_GetByID_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/units/u-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("u-1")

	svc := &mockUnitService{getResult: newTestUnit()}
	h := NewUnitHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestUnitHandler_GetByID_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/units/u-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("u-1")

	svc := &mockUnitService{getErr: application.ErrNotFound}
	h := NewUnitHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestUnitHandler_GetByID_Internal(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/units/u-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("u-1")

	svc := &mockUnitService{getErr: errors.New("db error")}
	h := NewUnitHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestUnitHandler_Update_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Room 202","status":"available"}`
	req := httptest.NewRequest(http.MethodPut, "/units/u-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("u-1")

	svc := &mockUnitService{}
	h := NewUnitHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestUnitHandler_Update_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid}`
	req := httptest.NewRequest(http.MethodPut, "/units/123", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewUnitHandler(&mockUnitService{})
	_ = h.Update(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestUnitHandler_Update_NotFound(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Room 202","status":"available"}`
	req := httptest.NewRequest(http.MethodPut, "/units/u-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("u-1")

	svc := &mockUnitService{updateErr: application.ErrNotFound}
	h := NewUnitHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestUnitHandler_Update_Internal(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Room 202","status":"available"}`
	req := httptest.NewRequest(http.MethodPut, "/units/u-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("u-1")

	svc := &mockUnitService{updateErr: errors.New("db error")}
	h := NewUnitHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestUnitHandler_Delete_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/units/u-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("u-1")

	svc := &mockUnitService{}
	h := NewUnitHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestUnitHandler_Delete_Internal(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/units/u-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("u-1")

	svc := &mockUnitService{deleteErr: errors.New("db error")}
	h := NewUnitHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestUnitHandler_NewUnitHandler(t *testing.T) {
	svc := &mockUnitService{}
	h := NewUnitHandler(svc)
	if h == nil || h.service == nil {
		t.Fatal("handler or service nil")
	}
}
