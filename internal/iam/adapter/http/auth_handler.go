package http

import (
	"errors"
	"github.com/ecelayes/pms-backend/internal/iam/application"
	"github.com/labstack/echo/v4"
	"net/http"
)

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	token, err := h.service.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"token": token})
}
func (h *AuthHandler) ForgotPassword(c echo.Context) error {
	var req ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	if err := h.service.RequestPasswordReset(c.Request().Context(), req.Email); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not process request"})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"message": "If the email exists, a reset link has been sent to your inbox.",
	})
}
func (h *AuthHandler) ResetPassword(c echo.Context) error {
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	if err := h.service.ResetPassword(c.Request().Context(), req.Token, req.NewPassword); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "password updated successfully"})
}

// Verify interface compliance
var _ AuthService = (*application.AuthService)(nil)
