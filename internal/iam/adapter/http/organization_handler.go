package http

import (
	"errors"
	"github.com/ecelayes/pms-backend/internal/iam/application"
	"github.com/ecelayes/pms-backend/internal/iam/domain"
	"github.com/ecelayes/pms-backend/internal/shared/dto"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

type OrganizationHandler struct {
	service OrganizationService
}

func NewOrganizationHandler(service OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{service: service}
}

type CreateOrganizationRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}
type UpdateOrganizationRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

func (h *OrganizationHandler) Create(c echo.Context) error {
	var req CreateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	id, err := h.service.Create(c.Request().Context(), req.Name, req.Code)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.JSON(http.StatusConflict, map[string]string{"error": "organization already exists"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"organization_id": id})
}
func (h *OrganizationHandler) GetAll(c echo.Context) error {
	orgs, err := h.service.GetAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	dtos := make([]dto.Organization, len(orgs))
	for i, o := range orgs {
		dtos[i] = h.toOrganizationDTO(o)
	}
	limit := len(orgs)
	if limit == 0 {
		limit = 10
	}
	response := dto.PaginatedResponse[dto.Organization]{
		Data: dtos,
		Meta: dto.Meta{
			TotalItems: int64(len(orgs)),
			TotalPages: 1,
			Page:       1,
			Limit:      limit,
		},
	}
	return c.JSON(http.StatusOK, response)
}
func (h *OrganizationHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	org, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, application.ErrOrgNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "organization not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.toOrganizationDTO(org))
}
func (h *OrganizationHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req UpdateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	if err := h.service.Update(c.Request().Context(), id, req.Name, req.Code); err != nil {
		if errors.Is(err, application.ErrOrgNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "organization not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "organization updated"})
}
func (h *OrganizationHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		if errors.Is(err, application.ErrOrgNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "organization not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "organization deleted"})
}
func (h *OrganizationHandler) toOrganizationDTO(o *domain.Organization) dto.Organization {
	return dto.Organization{
		ID:   o.ID(),
		Name: o.Name(),
		Code: o.Code(),
	}
}

// Verify interface compliance
var _ OrganizationService = (*application.OrganizationService)(nil)
