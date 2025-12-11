package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

type RatePlanHandler struct {
	uc *usecase.RatePlanUseCase
}

func NewRatePlanHandler(uc *usecase.RatePlanUseCase) *RatePlanHandler {
	return &RatePlanHandler{uc: uc}
}

func (h *RatePlanHandler) Create(c echo.Context) error {
	var req entity.CreateRatePlanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	id, err := h.uc.Create(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"rate_plan_id": id})
}

func (h *RatePlanHandler) List(c echo.Context) error {
	hotelID := c.QueryParam("hotel_id")
	if hotelID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "hotel_id is required"})
	}

	plans, err := h.uc.ListByHotel(c.Request().Context(), hotelID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	
	if plans == nil {
		plans = []entity.RatePlan{}
	}

	return c.JSON(http.StatusOK, plans)
}

func (h *RatePlanHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	plan, err := h.uc.GetByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "rate plan not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, plan)
}

func (h *RatePlanHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req entity.UpdateRatePlanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	if err := h.uc.Update(c.Request().Context(), id, req); err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "rate plan not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "rate plan updated"})
}

func (h *RatePlanHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.uc.Delete(c.Request().Context(), id); err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "rate plan not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "rate plan deleted"})
}
