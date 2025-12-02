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
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	id, err := h.uc.Create(c.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidDateRange), 
		     errors.Is(err, entity.ErrInvalidDateFormat):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			
		case errors.Is(err, entity.ErrNoAvailability), 
		     errors.Is(err, entity.ErrPriceChanged):
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})

		case errors.Is(err, entity.ErrRoomTypeNotFound):
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
			
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
	}

	return c.JSON(http.StatusCreated, map[string]string{"reservation_id": id})
}
