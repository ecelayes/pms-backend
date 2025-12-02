package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

type AvailabilityHandler struct {
	useCase *usecase.AvailabilityUseCase
}

func NewAvailabilityHandler(uc *usecase.AvailabilityUseCase) *AvailabilityHandler {
	return &AvailabilityHandler{useCase: uc}
}

func (h *AvailabilityHandler) Get(c echo.Context) error {
	startStr := c.QueryParam("start")
	endStr := c.QueryParam("end")

	layout := "2006-01-02"
	start, err := time.Parse(layout, startStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid start date (YYYY-MM-DD)"})
	}
	end, err := time.Parse(layout, endStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid end date (YYYY-MM-DD)"})
	}

	if !end.After(start) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "End date must be after start date"})
	}

	result, err := h.useCase.Search(c.Request().Context(), start, end)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"data": result})
}
