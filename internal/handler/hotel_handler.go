package handler

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

type HotelHandler struct {
	uc *usecase.HotelUseCase
}

func NewHotelHandler(uc *usecase.HotelUseCase) *HotelHandler {
	return &HotelHandler{uc: uc}
}

func (h *HotelHandler) Create(c echo.Context) error {
	var req entity.CreateHotelRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	ownerID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user context"})
	}

	id, err := h.uc.Create(c.Request().Context(), ownerID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"hotel_id": id})
}
