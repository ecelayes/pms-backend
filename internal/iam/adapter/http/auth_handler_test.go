package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
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
