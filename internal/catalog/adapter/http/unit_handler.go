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

type UnitHandler struct {
	service *application.CatalogService
}

func NewUnitHandler(service *application.CatalogService) *UnitHandler {
	return &UnitHandler{service: service}
}
func (h *UnitHandler) Create(c echo.Context) error {
	var req dto.CreateUnitRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	id, err := h.service.CreateUnit(c.Request().Context(), req.PropertyID, req.UnitTypeID, req.Name)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.JSON(http.StatusConflict, map[string]string{"error": "unit with this name already exists"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"unit_id": id})
}
func (h *UnitHandler) GetAll(c echo.Context) error {
	propertyID := c.QueryParam("property_id")
	if propertyID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "property_id is required"})
	}
	page, limit := 1, 10
	if p, err := strconv.Atoi(c.QueryParam("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(c.QueryParam("limit")); err == nil && l > 0 {
		limit = l
	}
	units, totalCount, err := h.service.ListUnits(c.Request().Context(), propertyID, page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	dtos := make([]dto.Unit, len(units))
	for i, u := range units {
		dtos[i] = h.toUnitDTO(u)
	}
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
	response := dto.PaginatedResponse[dto.Unit]{
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
func (h *UnitHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	u, err := h.service.GetUnit(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "unit not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.toUnitDTO(u))
}
func (h *UnitHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req dto.UpdateUnitRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	if err := h.service.UpdateUnit(c.Request().Context(), id, req.Name, req.Status); err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "unit not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "unit updated"})
}
func (h *UnitHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.DeleteUnit(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "unit deleted"})
}
func (h *UnitHandler) toUnitDTO(u *domain.Unit) dto.Unit {
	return dto.Unit{
		ID:         u.ID(),
		PropertyID: u.PropertyID(),
		UnitTypeID: u.UnitTypeID(),
		Name:       u.Name(),
		Status:     string(u.Status()),
	}
}
