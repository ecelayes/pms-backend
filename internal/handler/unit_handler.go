package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

type UnitHandler struct {
	uc *usecase.UnitUseCase
}

func NewUnitHandler(uc *usecase.UnitUseCase) *UnitHandler {
	return &UnitHandler{uc: uc}
}

func (h *UnitHandler) Create(c echo.Context) error {
	var req entity.CreateUnitRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	id, err := h.uc.Create(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
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

	units, err := h.uc.ListByProperty(c.Request().Context(), propertyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if units == nil {
		units = []entity.Unit{}
	}

	return c.JSON(http.StatusOK, units)
}

func (h *UnitHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	u, err := h.uc.GetByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "unit not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, u)
}

func (h *UnitHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req entity.UpdateUnitRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	if err := h.uc.Update(c.Request().Context(), id, req); err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "unit not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "unit updated"})
}

func (h *UnitHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.uc.Delete(c.Request().Context(), id); err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "unit not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "unit deleted"})
}
