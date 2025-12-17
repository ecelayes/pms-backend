package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

type AvailabilityHandler struct {
	uc *usecase.AvailabilityUseCase
}

func NewAvailabilityHandler(uc *usecase.AvailabilityUseCase) *AvailabilityHandler {
	return &AvailabilityHandler{uc: uc}
}

func (h *AvailabilityHandler) Get(c echo.Context) error {
	startStr := c.QueryParam("start")
	endStr := c.QueryParam("end")
	hotelID := c.QueryParam("hotel_id")
	
	roomsStr := c.QueryParam("rooms")
	adultsStr := c.QueryParam("adults")
	childrenStr := c.QueryParam("children")

	if startStr == "" || endStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "start and end dates required"})
	}

	layout := "2006-01-02"
	start, err := time.Parse(layout, startStr)
	if err != nil { return c.JSON(http.StatusBadRequest, map[string]string{"error": entity.ErrInvalidDateFormat.Error()}) }
	end, err := time.Parse(layout, endStr)
	if err != nil { return c.JSON(http.StatusBadRequest, map[string]string{"error": entity.ErrInvalidDateFormat.Error()}) }

	rooms := 1
	if roomsStr != "" { rooms, _ = strconv.Atoi(roomsStr) }
	
	adults := 1
	if adultsStr != "" { adults, _ = strconv.Atoi(adultsStr) }
	
	children := 0
	if childrenStr != "" { children, _ = strconv.Atoi(childrenStr) }

	page := 1
	limit := 10
	
	pageStr := c.QueryParam("page")
	if pageStr != "" { 
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 { page = p }
	}
	
	limitStr := c.QueryParam("limit")
	if limitStr != "" { 
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 { limit = l }
	}

	filter := entity.AvailabilityFilter{
		Start:    start,
		End:      end,
		HotelID:  hotelID,
		Rooms:    rooms,
		Adults:   adults,
		Children: children,
		Page:     page,
		Limit:    limit,
	}

	results, total, err := h.uc.Search(c.Request().Context(), filter)
	
	if err != nil {
		if errors.Is(err, entity.ErrInvalidDateRange) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}

	if results == nil {
		results = []entity.AvailabilitySearch{}
	}

	totalPage := int(total) / limit
	if int(total)%limit != 0 {
		totalPage++
	}

	response := entity.PaginatedResponse[entity.AvailabilitySearch]{
		Data: results,
		Meta: entity.PaginationMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPage,
		},
	}

	return c.JSON(http.StatusOK, response)
}
