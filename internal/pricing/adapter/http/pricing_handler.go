package http

import (
	"errors"
	"github.com/ecelayes/pms-backend/internal/pricing/application"
	"github.com/ecelayes/pms-backend/internal/pricing/domain"
	"github.com/ecelayes/pms-backend/internal/shared/dto"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

type PricingHandler struct {
	service *application.PricingService
}

func NewPricingHandler(service *application.PricingService) *PricingHandler {
	return &PricingHandler{service: service}
}

type SetPriceRequest struct {
	UnitTypeID string  `json:"unit_type_id"`
	Start      string  `json:"start"`
	End        string  `json:"end"`
	Price      float64 `json:"price"`
	Currency   string  `json:"currency"`
}

func (h *PricingHandler) BulkUpdate(c echo.Context) error {
	var req SetPriceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	start, err := time.Parse("2006-01-02", req.Start)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid start date format"})
	}
	end, err := time.Parse("2006-01-02", req.End)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid end date format"})
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}
	if req.Price < 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "price cannot be negative"})
	}
	priceCents := int64(req.Price * 100)
	price := vo.NewMoney(priceCents, req.Currency)
	if err := h.service.SetPriceRule(c.Request().Context(), req.UnitTypeID, start, end, price); err != nil {
		if errors.Is(err, vo.ErrinvalidDateRange) || strings.Contains(err.Error(), vo.ErrinvalidDateRange.Error()) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "price rules updated"})
}
func (h *PricingHandler) GetRules(c echo.Context) error {
	unitTypeID := c.QueryParam("unit_type_id")
	propertyID := c.QueryParam("property_id")
	if unitTypeID == "" && propertyID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "unit_type_id or property_id is required"})
	}
	rules, err := h.service.GetRules(c.Request().Context(), propertyID, unitTypeID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if rules == nil {
		rules = []*domain.PriceRule{}
	}
	dtos := make([]dto.PriceRule, len(rules))
	for i, r := range rules {
		dtos[i] = dto.PriceRule{
			ID:         r.ID(),
			UnitTypeID: r.UnitTypeID(),
			Start:      r.DateRange().Start().Format("2006-01-02"),
			End:        r.DateRange().End().Format("2006-01-02"),
			Price:      float64(r.Price().Amount()) / 100.0,
			Currency:   r.Price().Currency(),
		}
	}
	resp := dto.PaginatedResponse[dto.PriceRule]{
		Data: dtos,
		Meta: dto.Meta{
			TotalItems: int64(len(dtos)),
		},
	}
	return c.JSON(http.StatusOK, resp)
}
func (h *PricingHandler) DeleteRule(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.DeleteRule(c.Request().Context(), id); err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "rule deleted"})
}
