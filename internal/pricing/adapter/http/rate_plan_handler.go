package http

import (
	"errors"
	"github.com/ecelayes/pms-backend/internal/pricing/application"
	"github.com/ecelayes/pms-backend/internal/pricing/domain"
	"github.com/ecelayes/pms-backend/internal/shared/dto"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

type RatePlanHandler struct {
	service *application.PricingService
}

func NewRatePlanHandler(service *application.PricingService) *RatePlanHandler {
	return &RatePlanHandler{service: service}
}

type MealPlanDTO struct {
	Included    bool    `json:"included"`
	PricePerPax float64 `json:"price_per_pax"`
	Type        int     `json:"type"`
}
type CreateRatePlanRequest struct {
	PropertyID         string                    `json:"property_id"`
	Name               string                    `json:"name"`
	MealPlan           MealPlanDTO               `json:"meal_plan"`
	CancellationPolicy domain.CancellationPolicy `json:"cancellation_policy"`
	PaymentPolicy      domain.PaymentPolicy      `json:"payment_policy"`
}
type UpdateRatePlanRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

func (h *RatePlanHandler) Create(c echo.Context) error {
	var req CreateRatePlanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	mealPlan := domain.MealPlan{
		Included:    req.MealPlan.Included,
		PricePerPax: int64(req.MealPlan.PricePerPax * 100),
		Type:        req.MealPlan.Type,
	}
	id, err := h.service.CreateRatePlan(
		c.Request().Context(),
		req.PropertyID, req.Name,
		mealPlan, req.CancellationPolicy, req.PaymentPolicy,
	)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
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
	plans, err := h.service.ListRatePlans(c.Request().Context(), propertyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	var dtos []dto.RatePlan
	for _, p := range plans {
		dtos = append(dtos, dto.RatePlan{
			ID:         p.ID(),
			PropertyID: p.PropertyID(),
			UnitTypeID: getString(p.UnitTypeID()),
			Name:       p.Name(),
		})
	}
	return c.JSON(http.StatusOK, dto.PaginatedResponse[dto.RatePlan]{
		Data: dtos,
		Meta: dto.Meta{
			TotalItems: int64(len(dtos)),
			TotalPages: 1,
			Page:       1,
			Limit:      len(dtos),
		},
	})
}
func (h *RatePlanHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	plan, err := h.service.GetRatePlan(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "rate plan not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	dto := dto.RatePlan{
		ID:         plan.ID(),
		PropertyID: plan.PropertyID(),
		UnitTypeID: getString(plan.UnitTypeID()),
		Name:       plan.Name(),
	}
	return c.JSON(http.StatusOK, dto)
}
func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
func (h *RatePlanHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req UpdateRatePlanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	if err := h.service.UpdateRatePlan(c.Request().Context(), id, req.Name, req.Description, req.Active); err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "rate plan not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "rate plan updated"})
}
func (h *RatePlanHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.DeleteRatePlan(c.Request().Context(), id); err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "rate plan not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "rate plan deleted"})
}
