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
		if errors.Is(err, entity.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		if err.Error() == "unit type not found" {
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
	unitTypeID := c.QueryParam("unit_type_id")
	propertyID := c.QueryParam("property_id")

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

	rules, total, err := h.uc.GetRules(c.Request().Context(), unitTypeID, propertyID, pagination)
	if err != nil {
		if err.Error() == "unit_type_id or property_id is required" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if rules == nil {
		rules = []entity.PriceRule{}
	}

	totalPage := int(total) / pagination.Limit
	if int(total)%pagination.Limit != 0 {
		totalPage++
	}

	response := entity.PaginatedResponse[entity.PriceRule]{
		Data: rules,
		Meta: entity.PaginationMeta{
			Page:       pagination.Page,
			Limit:      pagination.Limit,
			TotalItems: total,
			TotalPages: totalPage,
		},
	}

	return c.JSON(http.StatusOK, response)
}
