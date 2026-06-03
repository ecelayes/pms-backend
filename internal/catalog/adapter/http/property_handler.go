package http

import (
	"errors"
	"github.com/ecelayes/pms-backend/internal/catalog/application"
	"github.com/ecelayes/pms-backend/internal/catalog/domain"
	"github.com/ecelayes/pms-backend/internal/shared/dto"
	"github.com/labstack/echo/v4"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type PropertyHandler struct {
	service *application.CatalogService
}

func NewPropertyHandler(service *application.CatalogService) *PropertyHandler {
	return &PropertyHandler{service: service}
}

type UpdatePropertyRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
	Type string `json:"type"`
}

func (h *PropertyHandler) Create(c echo.Context) error {
	var req dto.CreatePropertyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	id, err := h.service.CreateProperty(c.Request().Context(), req.OrganizationID, req.Name, req.Code, req.Type)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.JSON(http.StatusConflict, map[string]string{"error": "property with this code already exists"})
		}
		if strings.Contains(err.Error(), "invalid") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"property_id": id})
}
func (h *PropertyHandler) GetAll(c echo.Context) error {
	orgID := c.QueryParam("organization_id")
	page, limit := 1, 10
	if p, err := strconv.Atoi(c.QueryParam("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(c.QueryParam("limit")); err == nil && l > 0 {
		limit = l
	}
	properties, totalCount, err := h.service.ListProperties(c.Request().Context(), orgID, page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	dtos := make([]dto.Property, len(properties))
	for i, p := range properties {
		dtos[i] = h.toPropertyDTO(p)
	}
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
	response := dto.PaginatedResponse[dto.Property]{
		Data: dtos,
		Meta: dto.Meta{
			TotalItems: totalCount,
			TotalPages: totalPages,
			Page:       page,
			Limit:      limit,
		},
	}
	return c.JSON(http.StatusOK, response)
}
func (h *PropertyHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	property, err := h.service.GetProperty(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "property not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.toPropertyDTO(property))
}
func (h *PropertyHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req UpdatePropertyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	if err := h.service.UpdateProperty(c.Request().Context(), id, req.Name, req.Code, req.Type); err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "property not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "property updated"})
}
func (h *PropertyHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.DeleteProperty(c.Request().Context(), id); err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "property not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "property deleted"})
}
func (h *PropertyHandler) toPropertyDTO(p *domain.Property) dto.Property {
	return dto.Property{
		ID:             p.ID(),
		OrganizationID: p.OrganizationID(),
		Name:           p.Name(),
		Code:           p.Code(),
		Type:           string(p.Type()),
	}
}
