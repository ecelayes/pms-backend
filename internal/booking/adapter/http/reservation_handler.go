package http

import (
	"errors"
	"github.com/ecelayes/pms-backend/internal/booking/application"
	"github.com/ecelayes/pms-backend/internal/booking/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
	sharedContext "github.com/ecelayes/pms-backend/internal/shared/context"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

type ReservationHandler struct {
	service BookingService
}

func NewReservationHandler(service BookingService) *ReservationHandler {
	return &ReservationHandler{service: service}
}

type CreateReservationRequest struct {
	UnitTypeID     string `json:"unit_type_id"`
	Start          string `json:"start"`
	End            string `json:"end"`
	GuestEmail     string `json:"guest_email"`
	GuestFirstName string `json:"guest_first_name"`
	GuestLastName  string `json:"guest_last_name"`
	GuestPhone     string `json:"guest_phone"`
	RatePlanID     string `json:"rate_plan_id"`
	Adults         int    `json:"adults"`
	Children       int    `json:"children"`
}

func (h *ReservationHandler) Create(c echo.Context) error {
	var req CreateReservationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	if req.UnitTypeID == "" || req.GuestEmail == "" || req.RatePlanID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing unit_type_id, guest_email or rate_plan_id"})
	}
	layout := "2006-01-02"
	start, err := time.Parse(layout, req.Start)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid start date"})
	}
	end, err := time.Parse(layout, req.End)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid end date"})
	}
	code, err := h.service.CreateReservation(
		sharedContext.WithRequestID(c.Request().Context(), sharedContext.RequestIDFromEcho(c)),
		req.UnitTypeID,
		req.RatePlanID,
		start, end,
		req.GuestEmail,
		req.GuestFirstName,
		req.GuestLastName,
		req.GuestPhone,
		req.Adults,
		req.Children,
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidReservationDates):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, vo.ErrinvalidDateRange) || strings.Contains(err.Error(), vo.ErrinvalidDateRange.Error()):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, application.ErrNoAvailability):
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
	return c.JSON(http.StatusCreated, map[string]string{"reservation_code": code})
}
func (h *ReservationHandler) GetByCode(c echo.Context) error {
	code := c.Param("code")
	res, err := h.service.GetReservationByCode(sharedContext.WithRequestID(c.Request().Context(), sharedContext.RequestIDFromEcho(c)), code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if res == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":               res.ID(),
		"reservation_code": res.ReservationCode(),
		"guest_email":      res.GuestEmail(),
		"guest_id":         res.GuestID(),
		"status":           res.Status(),
		"unit_type_id":     res.UnitTypeID(),
		"start":            res.DateRange().Start().Format("2006-01-02"),
		"end":              res.DateRange().End().Format("2006-01-02"),
		"price_amount":     float64(res.Price().Amount()) / 100.0,
		"price_currency":   res.Price().Currency(),
	})
}
func (h *ReservationHandler) PreviewCancel(c echo.Context) error {
	id := c.Param("id")
	amount, err := h.service.PreviewCancellation(sharedContext.WithRequestID(c.Request().Context(), sharedContext.RequestIDFromEcho(c)), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]float64{"penalty_amount": amount})
}
func (h *ReservationHandler) Cancel(c echo.Context) error {
	id := c.Param("id")
	err := h.service.CancelReservation(sharedContext.WithRequestID(c.Request().Context(), sharedContext.RequestIDFromEcho(c)), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusOK)
}
func (h *ReservationHandler) Delete(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}
