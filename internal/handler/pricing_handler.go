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
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json body"})
	}

	err := h.uc.CreateRule(c.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrPriceNegative),
			 errors.Is(err, entity.ErrPriorityNegative),
			 errors.Is(err, entity.ErrInvalidDateFormat),
			 errors.Is(err, entity.ErrInvalidDateRange):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		
		case errors.Is(err, entity.ErrRoomTypeNotFound):
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
			
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "price rule created"})
}
