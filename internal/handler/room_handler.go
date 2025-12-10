package handler

import (
	"net/http"
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

type RoomHandler struct {
	uc *usecase.RoomUseCase
}

func NewRoomHandler(uc *usecase.RoomUseCase) *RoomHandler {
	return &RoomHandler{uc: uc}
}

func (h *RoomHandler) Create(c echo.Context) error {
	var req entity.CreateRoomTypeRequest 
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	id, err := h.uc.Create(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"room_type_id": id})
}

func (h *RoomHandler) GetAll(c echo.Context) error {
	hotelID := c.QueryParam("hotel_id")
	if hotelID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "hotel_id query param is required"})
	}

	rooms, err := h.uc.ListByHotel(c.Request().Context(), hotelID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, rooms)
}

func (h *RoomHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	
	room, err := h.uc.GetByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "room type not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, room)
}

func (h *RoomHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req entity.UpdateRoomTypeRequest
	if err := c.Bind(&req); err != nil { return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"}) }
	if err := h.uc.Update(c.Request().Context(), id, req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "room updated"})
}

func (h *RoomHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.uc.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "room deleted"})
}
