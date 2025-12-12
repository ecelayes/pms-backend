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

func (h *PricingHandler) BulkUpdate(c echo.Context) error {
	var req entity.SetPriceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	if err := h.uc.BulkCreateRule(c.Request().Context(), req); err != nil {
		if err.Error() == "room type not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "inventory updated successfully"})
}

func (h *PricingHandler) DeleteRule(c echo.Context) error {
	id := c.Param("id")
	if err := h.uc.DeleteRule(c.Request().Context(), id); err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "price rule not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "price rule deleted"})
}

func (h *PricingHandler) GetRules(c echo.Context) error {
	roomTypeID := c.QueryParam("room_type_id")
	if roomTypeID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "room_type_id query param is required"})
	}

	rules, err := h.uc.GetRules(c.Request().Context(), roomTypeID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, rules)
}
