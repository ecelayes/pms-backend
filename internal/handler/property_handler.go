package handler

import (
	"net/http"
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

type PropertyHandler struct {
	uc *usecase.PropertyUseCase
}

func NewPropertyHandler(uc *usecase.PropertyUseCase) *PropertyHandler {
	return &PropertyHandler{uc: uc}
}

func (h *PropertyHandler) Create(c echo.Context) error {
	var req entity.CreatePropertyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	id, err := h.uc.Create(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		if errors.Is(err, entity.ErrConflict) {
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"property_id": id})
}

func (h *PropertyHandler) GetAll(c echo.Context) error {
	orgID := c.QueryParam("organization_id")
	
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

	properties, total, err := h.uc.ListByOrganization(c.Request().Context(), orgID, pagination)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	
	if properties == nil {
		properties = []entity.Property{}
	}

	totalPage := int(total) / pagination.Limit
	if int(total)%pagination.Limit != 0 {
		totalPage++
	}

	response := entity.PaginatedResponse[entity.Property]{
		Data: properties,
		Meta: entity.PaginationMeta{
			Page:       pagination.Page,
			Limit:      pagination.Limit,
			TotalItems: total,
			TotalPages: totalPage,
		},
	}

	return c.JSON(http.StatusOK, response)
}

func (h *PropertyHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	property, err := h.uc.GetByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "property not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, property)
}

func (h *PropertyHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req entity.UpdatePropertyRequest
	if err := c.Bind(&req); err != nil { return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"}) }
	if err := h.uc.Update(c.Request().Context(), id, req); err != nil {
		if errors.Is(err, entity.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "property not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "property updated"})
}

func (h *PropertyHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.uc.Delete(c.Request().Context(), id); err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "property not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "property deleted"})
}
