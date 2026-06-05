package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ecelayes/pms-backend/internal/iam/application"
	sharedHTTP "github.com/ecelayes/pms-backend/internal/shared/adapter/http"
	sharedContext "github.com/ecelayes/pms-backend/internal/shared/context"
	"github.com/ecelayes/pms-backend/internal/shared/dto"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service UserService
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

type CreateUserRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	Role           string `json:"role"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Phone          string `json:"phone"`
	OrganizationID string `json:"organization_id"`
}
type UpdateUserRequest struct {
	Role      string `json:"role"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

func (h *UserHandler) Create(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	callerRole, ok := c.Get("role").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	if (req.Role == "owner" || req.Role == "admin") && callerRole != "super_admin" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "only super_admin can create organization owners"})
	}
	if req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "password is required"})
	}
	if err := sharedHTTP.ValidateEmail(req.Email); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	id, err := h.service.Register(
		sharedContext.WithRequestID(c.Request().Context(), sharedContext.RequestIDFromEcho(c)),
		req.OrganizationID,
		req.Email, req.Password, req.Role,
		req.FirstName, req.LastName, req.Phone,
	)
	if err != nil {
		if errors.Is(err, application.ErrWeakPassword) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.JSON(http.StatusConflict, map[string]string{"error": "user already exists"})
		}
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "validation") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"user_id": id})
}
func (h *UserHandler) GetAll(c echo.Context) error {
	orgID := c.QueryParam("organization_id")
	if err := sharedHTTP.ValidateRequired("organization_id", orgID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := sharedHTTP.ValidateUUID(orgID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	users, err := h.service.GetAll(sharedContext.WithRequestID(c.Request().Context(), sharedContext.RequestIDFromEcho(c)), orgID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	dtos := make([]dto.User, 0, len(users))
	for _, u := range users {
		dtos = append(dtos, dto.User{
			ID:        u.ID(),
			Email:     u.Email(),
			FirstName: u.FirstName(),
			LastName:  u.LastName(),
			Role:      string(u.Role()),
		})
	}
	return c.JSON(http.StatusOK, dto.PaginatedResponse[dto.User]{
		Data: dtos,
		Meta: dto.Meta{
			TotalItems: int64(len(dtos)),
			TotalPages: 1,
			Page:       1,
			Limit:      len(dtos),
		},
	})
}
func (h *UserHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	if err := sharedHTTP.ValidateUUID(id); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	u, err := h.service.GetByID(sharedContext.WithRequestID(c.Request().Context(), sharedContext.RequestIDFromEcho(c)), id)
	if err != nil {
		if errors.Is(err, application.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, dto.User{
		ID:        u.ID(),
		Email:     u.Email(),
		FirstName: u.FirstName(),
		LastName:  u.LastName(),
		Role:      string(u.Role()),
	})
}
func (h *UserHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	err := h.service.Update(
		sharedContext.WithRequestID(c.Request().Context(), sharedContext.RequestIDFromEcho(c)), id,
		req.Role, req.FirstName, req.LastName, req.Phone,
	)
	if err != nil {
		if errors.Is(err, application.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "user updated"})
}
func (h *UserHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.Delete(sharedContext.WithRequestID(c.Request().Context(), sharedContext.RequestIDFromEcho(c)), id); err != nil {
		if errors.Is(err, application.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "user deleted"})
}

// Verify interface compliance
var _ UserService = (*application.UserService)(nil)
