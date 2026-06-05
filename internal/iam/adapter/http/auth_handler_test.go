package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/iam/application"
)

func TestLoginRequest_JSON(t *testing.T) {
	jsonStr := `{"email":"user@example.com","password":"secret123"}`
	var req LoginRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}
	if req.Email != "user@example.com" {
		t.Errorf("Expected email 'user@example.com', got '%s'", req.Email)
	}
	if req.Password != "secret123" {
		t.Errorf("Expected password 'secret123', got '%s'", req.Password)
	}
}

func TestForgotPasswordRequest_JSON(t *testing.T) {
	jsonStr := `{"email":"user@example.com"}`
	var req ForgotPasswordRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}
	if req.Email != "user@example.com" {
		t.Errorf("Expected email 'user@example.com', got '%s'", req.Email)
	}
}

func TestResetPasswordRequest_JSON(t *testing.T) {
	jsonStr := `{"token":"abc123","new_password":"newsecret456"}`
	var req ResetPasswordRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}
	if req.Token != "abc123" {
		t.Errorf("Expected token 'abc123', got '%s'", req.Token)
	}
	if req.NewPassword != "newsecret456" {
		t.Errorf("Expected new_password 'newsecret456', got '%s'", req.NewPassword)
	}
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AuthHandler{service: nil}
	_ = h.Login(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestAuthHandler_ForgotPassword_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AuthHandler{service: nil}
	_ = h.ForgotPassword(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestAuthHandler_ForgotPassword_InvalidEmail(t *testing.T) {
	e := echo.New()
	reqBody := `{"email":"not-an-email"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AuthHandler{service: nil}
	_ = h.ForgotPassword(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestAuthHandler_ResetPassword_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AuthHandler{service: nil}
	_ = h.ResetPassword(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"email":"test@test.com","password":"password"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AuthHandler{service: &mockAuthService{loginToken: "jwt-token"}}
	_ = h.Login(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "jwt-token") {
		t.Errorf("Expected token in body, got: %s", rec.Body.String())
	}
}

func TestAuthHandler_Login_InvalidEmail(t *testing.T) {
	e := echo.New()
	reqBody := `{"email":"not-an-email","password":"password"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AuthHandler{service: nil}
	_ = h.Login(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	e := echo.New()
	reqBody := `{"email":"test@test.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AuthHandler{service: &mockAuthService{loginErr: application.ErrInvalidCredentials}}
	_ = h.Login(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestAuthHandler_Login_ServiceError(t *testing.T) {
	e := echo.New()
	reqBody := `{"email":"test@test.com","password":"password"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AuthHandler{service: &mockAuthService{loginErr: errors.New("internal error")}}
	_ = h.Login(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

func TestAuthHandler_ForgotPassword_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"email":"test@test.com"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AuthHandler{service: &mockAuthService{}}
	_ = h.ForgotPassword(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestAuthHandler_ForgotPassword_ServiceError(t *testing.T) {
	e := echo.New()
	reqBody := `{"email":"test@test.com"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AuthHandler{service: &mockAuthService{resetErr: errors.New("email service error")}}
	_ = h.ForgotPassword(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

func TestAuthHandler_ResetPassword_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"token":"valid-token","new_password":"newpass"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AuthHandler{service: &mockAuthService{}}
	_ = h.ResetPassword(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestAuthHandler_ResetPassword_WeakPassword(t *testing.T) {
	e := echo.New()
	reqBody := `{"token":"abc","new_password":"weak"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AuthHandler{service: &mockAuthService{resetPasswordErr: application.ErrWeakPassword}}
	_ = h.ResetPassword(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestAuthHandler_ResetPassword_InvalidToken(t *testing.T) {
	e := echo.New()
	reqBody := `{"token":"bad","new_password":"Good.Pass1"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AuthHandler{service: &mockAuthService{resetPasswordErr: errors.New("invalid token")}}
	_ = h.ResetPassword(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestAuthHandler_ResetPassword_ServiceError(t *testing.T) {
	e := echo.New()
	reqBody := `{"token":"invalid-token","new_password":"newpass"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &AuthHandler{service: &mockAuthService{resetPasswordErr: errors.New("invalid token")}}
	_ = h.ResetPassword(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestNewAuthHandler(t *testing.T) {
	svc := &mockAuthService{}
	h := NewAuthHandler(svc)
	if h == nil || h.service == nil {
		t.Fatal("handler or service nil")
	}
}
