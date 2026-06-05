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

type UnitTypeHandler struct {
	service UnitTypeService
}

func NewUnitTypeHandler(service UnitTypeService) *UnitTypeHandler {
	return &UnitTypeHandler{service: service}
}
func (h *UnitTypeHandler) Create(c echo.Context) error {
	var req dto.CreateUnitTypeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	priceCents := int64(req.BasePrice * 100)
	currency := "USD"
	id, err := h.service.CreateUnitType(
		c.Request().Context(),
		req.PropertyID, req.Name, req.Code,
		req.TotalQuantity, priceCents, currency,
		req.MaxOccupancy, req.MaxAdults, req.MaxChildren,
		req.Amenities,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.JSON(http.StatusConflict, map[string]string{"error": "unit type with this code already exists"})
		}
		if strings.Contains(err.Error(), "invalid") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"unit_type_id": id})
}
func (h *UnitTypeHandler) GetAll(c echo.Context) error {
	propertyID := c.QueryParam("property_id")
	page, limit := 1, 10
	if p, err := strconv.Atoi(c.QueryParam("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(c.QueryParam("limit")); err == nil && l > 0 {
		limit = l
	}
	unitTypes, totalCount, err := h.service.ListUnitTypes(c.Request().Context(), propertyID, page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	dtos := make([]dto.UnitType, len(unitTypes))
	for i, ut := range unitTypes {
		dtos[i] = h.toUnitTypeDTO(ut)
	}
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
	response := dto.PaginatedResponse[dto.UnitType]{
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
func (h *UnitTypeHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	ut, err := h.service.GetUnitTypeEntity(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "unit type not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.toUnitTypeDTO(ut))
}
func (h *UnitTypeHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req struct {
		Name         string   `json:"name"`
		Code         string   `json:"code"`
		TotalQty     int      `json:"total_quantity"`
		BasePrice    float64  `json:"base_price"`
		MaxOccupancy int      `json:"max_occupancy"`
		MaxAdults    int      `json:"max_adults"`
		MaxChildren  int      `json:"max_children"`
		Amenities    []string `json:"amenities"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	priceCents := int64(req.BasePrice * 100)
	currency := "USD"
	err := h.service.UpdateUnitType(
		c.Request().Context(), id,
		req.Name, req.Code, req.TotalQty, priceCents, currency,
		req.MaxOccupancy, req.MaxAdults, req.MaxChildren, req.Amenities,
	)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "unit type not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "unit type updated"})
}
func (h *UnitTypeHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.DeleteUnitType(c.Request().Context(), id); err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "unit type not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "unit type deleted"})
}
func (h *UnitTypeHandler) toUnitTypeDTO(ut *domain.UnitType) dto.UnitType {
	return dto.UnitType{
		ID:           ut.ID(),
		PropertyID:   ut.PropertyID(),
		Name:         ut.Name(),
		Code:         ut.Code(),
		TotalQty:     ut.TotalQuantity(),
		BasePrice:    float64(ut.BasePrice().Amount()) / 100.0,
		MaxOccupancy: ut.MaxOccupancy(),
		Amenities:    ut.Amenities(),
	}
}
