package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ecelayes/pms-backend/internal/iam/application"
	"github.com/ecelayes/pms-backend/internal/iam/domain"
	"github.com/labstack/echo/v4"
)

func TestOrganizationHandler_Create_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Test Org","code":"TST"}`
	req := httptest.NewRequest(http.MethodPost, "/organizations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockOrganizationService{createResult: "org-123"}
	h := NewOrganizationHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rec.Code)
	}
}

func TestOrganizationHandler_Create_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/organizations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewOrganizationHandler(&mockOrganizationService{})
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestOrganizationHandler_Create_InvalidValidation(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Test Org","code":"TST"}`
	req := httptest.NewRequest(http.MethodPost, "/organizations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockOrganizationService{createErr: errors.New("invalid name")}
	h := NewOrganizationHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestOrganizationHandler_Create_Duplicate(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Test Org","code":"TST"}`
	req := httptest.NewRequest(http.MethodPost, "/organizations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockOrganizationService{createErr: errors.New("duplicate key violates unique constraint")}
	h := NewOrganizationHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", rec.Code)
	}
}

func TestOrganizationHandler_Create_InternalError(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Test Org","code":"TST"}`
	req := httptest.NewRequest(http.MethodPost, "/organizations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockOrganizationService{createErr: errors.New("db connection failed")}
	h := NewOrganizationHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

func TestOrganizationHandler_GetAll_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	org, _ := domain.NewOrganization("Org 1", "ORG1")
	svc := &mockOrganizationService{getAllResult: []*domain.Organization{org}}
	h := NewOrganizationHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestOrganizationHandler_GetAll_Empty(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockOrganizationService{getAllResult: []*domain.Organization{}}
	h := NewOrganizationHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Meta struct {
			Limit int `json:"limit"`
		} `json:"meta"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Meta.Limit != 10 {
		t.Errorf("Expected default limit 10, got %d", resp.Meta.Limit)
	}
}

func TestOrganizationHandler_GetAll_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockOrganizationService{getAllErr: errors.New("db error")}
	h := NewOrganizationHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

func TestOrganizationHandler_GetByID_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/organizations/o1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("o1")

	org, _ := domain.NewOrganization("Org 1", "ORG1")
	svc := &mockOrganizationService{getByIDResult: org}
	h := NewOrganizationHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestOrganizationHandler_GetByID_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/organizations/o1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("o1")

	svc := &mockOrganizationService{getByIDErr: application.ErrOrgNotFound}
	h := NewOrganizationHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}

func TestOrganizationHandler_GetByID_InternalError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/organizations/o1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("o1")

	svc := &mockOrganizationService{getByIDErr: errors.New("db error")}
	h := NewOrganizationHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

func TestOrganizationHandler_Update_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated","code":"UPD"}`
	req := httptest.NewRequest(http.MethodPut, "/organizations/o1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("o1")

	svc := &mockOrganizationService{}
	h := NewOrganizationHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestOrganizationHandler_Update_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPut, "/organizations/123", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewOrganizationHandler(&mockOrganizationService{})
	_ = h.Update(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestOrganizationHandler_Update_NotFound(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated","code":"UPD"}`
	req := httptest.NewRequest(http.MethodPut, "/organizations/o1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("o1")

	svc := &mockOrganizationService{updateErr: application.ErrOrgNotFound}
	h := NewOrganizationHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}

func TestOrganizationHandler_Update_Error(t *testing.T) {
	e := echo.New()
	reqBody := `{"name":"Updated","code":"UPD"}`
	req := httptest.NewRequest(http.MethodPut, "/organizations/o1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("o1")

	svc := &mockOrganizationService{updateErr: errors.New("db error")}
	h := NewOrganizationHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

func TestOrganizationHandler_Delete_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/organizations/o1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("o1")

	svc := &mockOrganizationService{}
	h := NewOrganizationHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestOrganizationHandler_Delete_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/organizations/o1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("o1")

	svc := &mockOrganizationService{deleteErr: application.ErrOrgNotFound}
	h := NewOrganizationHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}

func TestOrganizationHandler_Delete_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/organizations/o1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("o1")

	svc := &mockOrganizationService{deleteErr: errors.New("db error")}
	h := NewOrganizationHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

func TestCreateOrganizationRequest_JSON(t *testing.T) {
	jsonStr := `{"name":"Test Org","code":"TST"}`
	var req CreateOrganizationRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if req.Name != "Test Org" || req.Code != "TST" {
		t.Errorf("JSON unmarshal failed: name=%s code=%s", req.Name, req.Code)
	}
}

func TestUpdateOrganizationRequest_JSON(t *testing.T) {
	jsonStr := `{"name":"Updated Org","code":"UPD"}`
	var req UpdateOrganizationRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if req.Name != "Updated Org" || req.Code != "UPD" {
		t.Errorf("JSON unmarshal failed: name=%s code=%s", req.Name, req.Code)
	}
}

func TestOrganizationHandler_NewOrganizationHandler(t *testing.T) {
	svc := &mockOrganizationService{}
	h := NewOrganizationHandler(svc)
	if h == nil {
		t.Fatal("NewOrganizationHandler returned nil")
	}
	if h.service == nil {
		t.Fatal("service is nil")
	}
}

// keep context import warning-free
var _ = context.Background
