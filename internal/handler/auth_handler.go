package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

type AuthHandler struct {
	uc *usecase.AuthUseCase
}

func NewAuthHandler(uc *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req entity.AuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	token, err := h.uc.Login(c.Request().Context(), req)
	if err != nil {
		if err == entity.ErrInvalidCredentials {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

func (h *AuthHandler) ForgotPassword(c echo.Context) error {
	var req entity.ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	if err := h.uc.RequestPasswordReset(c.Request().Context(), req.Email); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not process request"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "If the email exists, a reset link has been sent to your inbox.",
	})
}

func (h *AuthHandler) ResetPassword(c echo.Context) error {
	var req entity.ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	if err := h.uc.ResetPassword(c.Request().Context(), req); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "password updated successfully"})
}
