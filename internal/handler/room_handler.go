package handler

import (
	"net/http"
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
	var req usecase.CreateRoomTypeRequest
	if err := c.Bind(&req); err != nil { return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"}) }
	id, err := h.uc.Create(c.Request().Context(), req)
	if err != nil { return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()}) }
	return c.JSON(http.StatusCreated, map[string]string{"room_type_id": id})
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
