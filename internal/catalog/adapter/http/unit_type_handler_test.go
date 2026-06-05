package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ecelayes/pms-backend/internal/catalog/application"
	"github.com/ecelayes/pms-backend/internal/catalog/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"github.com/labstack/echo/v4"
)

func TestUnitTypeHandler_Create_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"property_id":"p-1","name":"Deluxe","code":"DLX","total_quantity":5,"base_price":100.0,"max_occupancy":2,"max_adults":2,"max_children":1,"amenities":["wifi"]}`
	req := httptest.NewRequest(http.MethodPost, "/unit-types", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockUnitTypeService{createResult: "ut-1"}
	h := NewUnitTypeHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_Create_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid}`
	req := httptest.NewRequest(http.MethodPost, "/unit-types", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewUnitTypeHandler(&mockUnitTypeService{})
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_Create_Duplicate(t *testing.T) {
	e := echo.New()
	reqBody := `{"property_id":"p-1","name":"Deluxe","code":"DLX","total_quantity":5,"base_price":100.0,"max_occupancy":2,"max_adults":2,"max_children":1,"amenities":[]}`
	req := httptest.NewRequest(http.MethodPost, "/unit-types", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockUnitTypeService{createErr: errors.New("duplicate key")}
	h := NewUnitTypeHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusConflict {
		t.Errorf("Expected 409, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_Create_Invalid(t *testing.T) {
	e := echo.New()
	reqBody := `{"property_id":"p-1","name":"Deluxe","code":"DLX","total_quantity":5,"base_price":100.0,"max_occupancy":2,"max_adults":2,"max_children":1,"amenities":[]}`
	req := httptest.NewRequest(http.MethodPost, "/unit-types", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockUnitTypeService{createErr: errors.New("invalid input")}
	h := NewUnitTypeHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_Create_Internal(t *testing.T) {
	e := echo.New()
	reqBody := `{"property_id":"p-1","name":"Deluxe","code":"DLX","total_quantity":5,"base_price":100.0,"max_occupancy":2,"max_adults":2,"max_children":1,"amenities":[]}`
	req := httptest.NewRequest(http.MethodPost, "/unit-types", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockUnitTypeService{createErr: errors.New("db error")}
	h := NewUnitTypeHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_GetAll_Default(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/unit-types", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockUnitTypeService{listResult: nil, listTotal: 0}
	h := NewUnitTypeHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_GetAll_WithParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/unit-types?property_id=p-1&page=1&limit=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockUnitTypeService{listResult: nil, listTotal: 0}
	h := NewUnitTypeHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_GetAll_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/unit-types", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockUnitTypeService{listErr: errors.New("db error")}
	h := NewUnitTypeHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_GetByID_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/unit-types/ut-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("ut-1")

	svc := &mockUnitTypeService{getErr: application.ErrNotFound}
	h := NewUnitTypeHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_GetByID_Internal(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/unit-types/ut-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("ut-1")

	svc := &mockUnitTypeService{getErr: errors.New("db error")}
	h := NewUnitTypeHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_Update_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated","code":"UPD","total_quantity":10,"base_price":150.0,"max_occupancy":2,"max_adults":2,"max_children":1,"amenities":["wifi"]}`
	req := httptest.NewRequest(http.MethodPut, "/unit-types/ut-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("ut-1")

	svc := &mockUnitTypeService{}
	h := NewUnitTypeHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_Update_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid}`
	req := httptest.NewRequest(http.MethodPut, "/unit-types/ut-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewUnitTypeHandler(&mockUnitTypeService{})
	_ = h.Update(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_Update_NotFound(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated","code":"UPD","total_quantity":10,"base_price":150.0,"max_occupancy":2,"max_adults":2,"max_children":1,"amenities":[]}`
	req := httptest.NewRequest(http.MethodPut, "/unit-types/ut-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("ut-1")

	svc := &mockUnitTypeService{updateErr: application.ErrNotFound}
	h := NewUnitTypeHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_Update_Internal(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated","code":"UPD","total_quantity":10,"base_price":150.0,"max_occupancy":2,"max_adults":2,"max_children":1,"amenities":[]}`
	req := httptest.NewRequest(http.MethodPut, "/unit-types/ut-1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("ut-1")

	svc := &mockUnitTypeService{updateErr: errors.New("db error")}
	h := NewUnitTypeHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_Delete_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/unit-types/ut-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("ut-1")

	svc := &mockUnitTypeService{}
	h := NewUnitTypeHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_Delete_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/unit-types/ut-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("ut-1")

	svc := &mockUnitTypeService{deleteErr: application.ErrNotFound}
	h := NewUnitTypeHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_Delete_Internal(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/unit-types/ut-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("ut-1")

	svc := &mockUnitTypeService{deleteErr: errors.New("db error")}
	h := NewUnitTypeHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_NewUnitTypeHandler(t *testing.T) {
	svc := &mockUnitTypeService{}
	h := NewUnitTypeHandler(svc)
	if h == nil || h.service == nil {
		t.Fatal("handler or service nil")
	}
}

func TestUnitTypeHandler_GetByID_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/unit-types/ut-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("ut-1")

	ut, _ := domain.NewUnitType("p-1", "Standard", "STD", 5, vo.NewMoney(10000, "USD"), 2, 2, 0, []string{"wifi"})
	svc := &mockUnitTypeService{getResult: ut}
	h := NewUnitTypeHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestUnitTypeHandler_toUnitTypeDTO(t *testing.T) {
	ut, _ := domain.NewUnitType("p-1", "Deluxe", "DLX", 3, vo.NewMoney(15000, "USD"), 4, 2, 2, []string{"wifi", "tv"})
	h := NewUnitTypeHandler(&mockUnitTypeService{})
	d := h.toUnitTypeDTO(ut)
	if d.ID != ut.ID() || d.PropertyID != "p-1" || d.Name != "Deluxe" || d.Code != "DLX" {
		t.Errorf("DTO mismatch: %+v", d)
	}
	if d.BasePrice != 150.0 {
		t.Errorf("BasePrice = %f, want 150.0", d.BasePrice)
	}
	if d.MaxOccupancy != 4 {
		t.Errorf("MaxOccupancy = %d, want 4", d.MaxOccupancy)
	}
	if len(d.Amenities) != 2 {
		t.Errorf("Amenities = %v, want 2 items", d.Amenities)
	}
}

func TestUnitTypeHandler_GetAll_WithData(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/unit-types?property_id=p-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ut, _ := domain.NewUnitType("p-1", "Standard", "STD", 5, vo.NewMoney(10000, "USD"), 2, 2, 0, []string{"wifi"})
	svc := &mockUnitTypeService{listResult: []*domain.UnitType{ut}, listTotal: 1}
	h := NewUnitTypeHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}
