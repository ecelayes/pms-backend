package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

type AvailabilityHandler struct {
	uc *usecase.AvailabilityUseCase
}

func NewAvailabilityHandler(uc *usecase.AvailabilityUseCase) *AvailabilityHandler {
	return &AvailabilityHandler{uc: uc}
}

func (h *AvailabilityHandler) Get(c echo.Context) error {
	startStr := c.QueryParam("start")
	endStr := c.QueryParam("end")

	if startStr == "" || endStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "start and end dates are required"})
	}

	layout := "2006-01-02"
	start, err := time.Parse(layout, startStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": entity.ErrInvalidDateFormat.Error()})
	}
	end, err := time.Parse(layout, endStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": entity.ErrInvalidDateFormat.Error()})
	}

	results, err := h.uc.Search(c.Request().Context(), start, end)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidDateRange) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	if results == nil {
		results = []entity.AvailabilitySearch{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"data": results})
}
