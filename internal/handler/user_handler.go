package handler

import (
	"net/http"
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

type UserHandler struct {
	uc *usecase.UserUseCase
}

func NewUserHandler(uc *usecase.UserUseCase) *UserHandler {
	return &UserHandler{uc: uc}
}

func (h *UserHandler) Create(c echo.Context) error {
	requesterRole, ok := c.Get("role").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing role information"})
	}

	var req entity.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	id, err := h.uc.CreateUser(c.Request().Context(), requesterRole, req)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		if errors.Is(err, entity.ErrConflict) {
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}
		if err.Error() == "only super_admin can create organization owners" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"user_id": id})
}

func (h *UserHandler) GetAll(c echo.Context) error {
	orgID := c.QueryParam("organization_id")
	if orgID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "organization_id query param required"})
	}

	var pagination entity.PaginationRequest
	if err := c.Bind(&pagination); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid pagination params"})
	}
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 {
		pagination.Limit = 10 
	}

	users, total, err := h.uc.GetAll(c.Request().Context(), orgID, pagination)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	
	if users == nil {
		users = []entity.User{}
	}

	totalPage := int(total) / pagination.Limit
	if int(total)%pagination.Limit != 0 {
		totalPage++
	}

	response := entity.PaginatedResponse[entity.User]{
		Data: users,
		Meta: entity.PaginationMeta{
			Page:       pagination.Page,
			Limit:      pagination.Limit,
			TotalItems: total,
			TotalPages: totalPage,
		},
	}

	return c.JSON(http.StatusOK, response)
}

func (h *UserHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	user, err := h.uc.GetByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Update(c echo.Context) error {
	id := c.Param("id")
	orgID := c.QueryParam("organization_id") 
	if orgID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "organization_id query param required for role update"})
	}

	var req entity.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	err := h.uc.Update(c.Request().Context(), id, orgID, req)
	if err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found in this organization"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "user updated"})
}

func (h *UserHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	err := h.uc.Delete(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "user deleted"})
}
