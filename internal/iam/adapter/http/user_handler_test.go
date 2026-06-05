package http

import (
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

func TestUserHandler_Create_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", "super_admin")

	h := NewUserHandler(&mockUserService{})
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestUserHandler_Create_Unauthorized(t *testing.T) {
	e := echo.New()
	reqBody := `{"email":"test@test.com","password":"pass123","role":"admin"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", nil)

	h := NewUserHandler(&mockUserService{})
	_ = h.Create(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestUserHandler_Create_ForbiddenOwner(t *testing.T) {
	e := echo.New()
	reqBody := `{"email":"test@test.com","password":"pass123","role":"owner"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", "admin")

	h := NewUserHandler(&mockUserService{})
	_ = h.Create(c)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rec.Code)
	}
}

func TestUserHandler_Create_ForbiddenAdmin(t *testing.T) {
	e := echo.New()
	reqBody := `{"email":"test@test.com","password":"pass123","role":"admin"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", "manager")

	h := NewUserHandler(&mockUserService{})
	_ = h.Create(c)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rec.Code)
	}
}

func TestUserHandler_Create_MissingPassword(t *testing.T) {
	e := echo.New()
	reqBody := `{"email":"test@test.com","role":"admin"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", "super_admin")

	h := NewUserHandler(&mockUserService{})
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestUserHandler_Create_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"email":"test@test.com","password":"pass123","role":"manager","organization_id":"o1"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", "super_admin")

	svc := &mockUserService{registerResult: "user-123"}
	h := NewUserHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rec.Code)
	}
}

func TestUserHandler_Create_Duplicate(t *testing.T) {
	e := echo.New()
	reqBody := `{"email":"test@test.com","password":"pass123","role":"manager","organization_id":"o1"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", "super_admin")

	svc := &mockUserService{registerErr: errors.New("duplicate key violates unique constraint")}
	h := NewUserHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", rec.Code)
	}
}

func TestUserHandler_Create_Validation(t *testing.T) {
	e := echo.New()
	reqBody := `{"email":"test@test.com","password":"pass123","role":"manager","organization_id":"o1"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", "super_admin")

	svc := &mockUserService{registerErr: errors.New("invalid email format")}
	h := NewUserHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestUserHandler_Create_InternalError(t *testing.T) {
	e := echo.New()
	reqBody := `{"email":"test@test.com","password":"pass123","role":"manager","organization_id":"o1"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", "super_admin")

	svc := &mockUserService{registerErr: errors.New("db error")}
	h := NewUserHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

func TestUserHandler_GetAll_MissingOrgID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewUserHandler(&mockUserService{})
	_ = h.GetAll(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestUserHandler_GetAll_Empty(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users?organization_id=550e8400-e29b-41d4-a716-446655440000", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockUserService{getAllResult: nil}
	h := NewUserHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestUserHandler_GetAll_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users?organization_id=550e8400-e29b-41d4-a716-446655440000", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	user, _ := domain.NewUser("test@test.com", "hashed", "salt", "manager", "John", "Doe", "123")
	svc := &mockUserService{getAllResult: []*domain.User{user}}
	h := NewUserHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestUserHandler_GetAll_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users?organization_id=550e8400-e29b-41d4-a716-446655440000", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockUserService{getAllErr: errors.New("db error")}
	h := NewUserHandler(svc)
	_ = h.GetAll(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

func TestUserHandler_GetByID_InvalidUUID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	h := NewUserHandler(&mockUserService{})
	err := h.GetByID(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestUserHandler_GetByID_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/550e8400-e29b-41d4-a716-446655440000", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("550e8400-e29b-41d4-a716-446655440000")

	user, _ := domain.NewUser("test@test.com", "hashed", "salt", "manager", "John", "Doe", "123")
	svc := &mockUserService{getByIDResult: user}
	h := NewUserHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/550e8400-e29b-41d4-a716-446655440000", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("550e8400-e29b-41d4-a716-446655440000")

	svc := &mockUserService{getByIDErr: application.ErrUserNotFound}
	h := NewUserHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}

func TestUserHandler_GetByID_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/550e8400-e29b-41d4-a716-446655440000", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("550e8400-e29b-41d4-a716-446655440000")

	svc := &mockUserService{getByIDErr: errors.New("db error")}
	h := NewUserHandler(svc)
	_ = h.GetByID(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

func TestUserHandler_Update_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"role":"manager","first_name":"Jane","last_name":"Smith","phone":"456"}`
	req := httptest.NewRequest(http.MethodPut, "/users/550e8400-e29b-41d4-a716-446655440000", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("550e8400-e29b-41d4-a716-446655440000")

	svc := &mockUserService{}
	h := NewUserHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestUserHandler_Update_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPut, "/users/123", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewUserHandler(&mockUserService{})
	_ = h.Update(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestUserHandler_Update_NotFound(t *testing.T) {
	e := echo.New()
	reqBody := `{"role":"manager","first_name":"Jane","last_name":"Smith","phone":"456"}`
	req := httptest.NewRequest(http.MethodPut, "/users/550e8400-e29b-41d4-a716-446655440000", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("550e8400-e29b-41d4-a716-446655440000")

	svc := &mockUserService{updateErr: application.ErrUserNotFound}
	h := NewUserHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}

func TestUserHandler_Update_Error(t *testing.T) {
	e := echo.New()
	reqBody := `{"role":"manager","first_name":"Jane","last_name":"Smith","phone":"456"}`
	req := httptest.NewRequest(http.MethodPut, "/users/550e8400-e29b-41d4-a716-446655440000", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("550e8400-e29b-41d4-a716-446655440000")

	svc := &mockUserService{updateErr: errors.New("db error")}
	h := NewUserHandler(svc)
	_ = h.Update(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

func TestUserHandler_Delete_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/users/550e8400-e29b-41d4-a716-446655440000", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("550e8400-e29b-41d4-a716-446655440000")

	svc := &mockUserService{}
	h := NewUserHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestUserHandler_Delete_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/users/550e8400-e29b-41d4-a716-446655440000", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("550e8400-e29b-41d4-a716-446655440000")

	svc := &mockUserService{deleteErr: application.ErrUserNotFound}
	h := NewUserHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}

func TestUserHandler_Delete_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/users/550e8400-e29b-41d4-a716-446655440000", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("550e8400-e29b-41d4-a716-446655440000")

	svc := &mockUserService{deleteErr: errors.New("db error")}
	h := NewUserHandler(svc)
	_ = h.Delete(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

func TestCreateUserRequest_JSON(t *testing.T) {
	jsonStr := `{"email":"test@test.com","password":"secret123","role":"admin","first_name":"John","last_name":"Doe","phone":"123","organization_id":"org-123"}`
	var req CreateUserRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if req.Email != "test@test.com" || req.Password != "secret123" || req.Role != "admin" {
		t.Errorf("JSON unmarshal failed: %+v", req)
	}
}

func TestUpdateUserRequest_JSON(t *testing.T) {
	jsonStr := `{"role":"owner","first_name":"Jane","last_name":"Smith","phone":"456"}`
	var req UpdateUserRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if req.Role != "owner" || req.FirstName != "Jane" {
		t.Errorf("JSON unmarshal failed: %+v", req)
	}
}

func TestUserHandler_NewUserHandler(t *testing.T) {
	svc := &mockUserService{}
	h := NewUserHandler(svc)
	if h == nil {
		t.Fatal("NewUserHandler returned nil")
	}
	if h.service == nil {
		t.Fatal("service is nil")
	}
}
