package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

type OrganizationHandler struct {
	uc *usecase.OrganizationUseCase
}

func NewOrganizationHandler(uc *usecase.OrganizationUseCase) *OrganizationHandler {
	return &OrganizationHandler{uc: uc}
}

func (h *OrganizationHandler) Create(c echo.Context) error {
	var req entity.CreateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	id, err := h.uc.Create(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"organization_id": id})
}

func (h *OrganizationHandler) GetAll(c echo.Context) error {
	orgs, err := h.uc.GetAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, orgs)
}

func (h *OrganizationHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	org, err := h.uc.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, org)
}

func (h *OrganizationHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req entity.UpdateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	if err := h.uc.Update(c.Request().Context(), id, req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "updated successfully"})
}

func (h *OrganizationHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.uc.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted successfully"})
}
