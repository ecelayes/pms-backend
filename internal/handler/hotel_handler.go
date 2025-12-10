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

	if req.OrganizationID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "organization_id is required"})
	}

	id, err := h.uc.Create(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"hotel_id": id})
}

func (h *HotelHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req entity.UpdateHotelRequest
	if err := c.Bind(&req); err != nil { return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"}) }
	if err := h.uc.Update(c.Request().Context(), id, req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "hotel updated"})
}

func (h *HotelHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.uc.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "hotel deleted"})
}

func (h *HotelHandler) GetAll(c echo.Context) error {
	orgID := c.QueryParam("organization_id")
	
	if orgID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "organization_id query param is required"})
	}

	hotels, err := h.uc.ListByOrganization(c.Request().Context(), orgID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"data": hotels})
}
