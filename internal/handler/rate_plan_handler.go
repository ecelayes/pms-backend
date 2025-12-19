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
		if errors.Is(err, entity.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"rate_plan_id": id})
}

func (h *RatePlanHandler) List(c echo.Context) error {
	propertyID := c.QueryParam("property_id")
	if propertyID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "property_id is required"})
	}

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

	plans, total, err := h.uc.ListByProperty(c.Request().Context(), propertyID, pagination)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	
	if plans == nil {
		plans = []entity.RatePlan{}
	}

	totalPage := int(total) / pagination.Limit
	if int(total)%pagination.Limit != 0 {
		totalPage++
	}

	response := entity.PaginatedResponse[entity.RatePlan]{
		Data: plans,
		Meta: entity.PaginationMeta{
			Page:       pagination.Page,
			Limit:      pagination.Limit,
			TotalItems: total,
			TotalPages: totalPage,
		},
	}

	return c.JSON(http.StatusOK, response)
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
