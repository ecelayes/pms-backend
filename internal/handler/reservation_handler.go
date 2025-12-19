package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

type ReservationHandler struct {
	uc *usecase.ReservationUseCase
}

func NewReservationHandler(uc *usecase.ReservationUseCase) *ReservationHandler {
	return &ReservationHandler{uc: uc}
}

func (h *ReservationHandler) Create(c echo.Context) error {
	var req entity.CreateReservationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	if req.UnitTypeID == "" || req.GuestEmail == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing unit_type_id or guest_email"})
	}
	if req.GuestFirstName == "" || req.GuestLastName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "guest name is required"})
	}

	code, err := h.uc.Create(c.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidDateFormat), 
		     errors.Is(err, entity.ErrInvalidDateRange),
		     errors.Is(err, entity.ErrInvalidInput):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, entity.ErrNoAvailability):
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		case errors.Is(err, entity.ErrUnitTypeNotFound):
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}

	return c.JSON(http.StatusCreated, map[string]string{"reservation_code": code})
}

func (h *ReservationHandler) GetByCode(c echo.Context) error {
	code := c.Param("code")
	res, err := h.uc.GetByCode(c.Request().Context(), code)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "reservation not found"})
	}
	return c.JSON(http.StatusOK, res)
}

func (h *ReservationHandler) PreviewCancel(c echo.Context) error {
	id := c.Param("id")
	
	penalty, err := h.uc.PreviewCancellation(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "reservation not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"penalty_amount": penalty,
		"currency":       "USD",
	})
}

func (h *ReservationHandler) Cancel(c echo.Context) error {
	id := c.Param("id") // TODO: Here it is still by internal UUID for operations, or it could be by code.
	err := h.uc.Cancel(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "cancelled"})
}

func (h *ReservationHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.uc.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "reservation deleted"})
}
