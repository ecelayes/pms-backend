package handler

import (
	"errors"
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

type PricingHandler struct {
	uc *usecase.PricingUseCase
}

func NewPricingHandler(uc *usecase.PricingUseCase) *PricingHandler {
	return &PricingHandler{uc: uc}
}

func (h *PricingHandler) CreateRule(c echo.Context) error {
	var req entity.CreatePriceRuleRequest
	if err := c.Bind(&req); err != nil { return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"}) }
	if err := h.uc.CreateRule(c.Request().Context(), req); err != nil {
		if errors.Is(err, entity.ErrRoomTypeNotFound) { return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()}) }
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"message": "rule created"})
}

func (h *PricingHandler) UpdateRule(c echo.Context) error {
	id := c.Param("id")
	var req entity.UpdatePriceRuleRequest
	if err := c.Bind(&req); err != nil { return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"}) }
	if err := h.uc.UpdateRule(c.Request().Context(), id, req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "rule updated"})
}

func (h *PricingHandler) DeleteRule(c echo.Context) error {
	id := c.Param("id")
	if err := h.uc.DeleteRule(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "rule deleted"})
}
