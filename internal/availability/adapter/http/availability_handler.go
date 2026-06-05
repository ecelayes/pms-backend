package http

import (
	sharedContext "github.com/ecelayes/pms-backend/internal/shared/context"
	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/shared/dto"
	"math"
	"net/http"
	"strconv"
	"time"
)

type AvailabilityHandler struct {
	service AvailabilityService
}

func NewAvailabilityHandler(service AvailabilityService) *AvailabilityHandler {
	return &AvailabilityHandler{service: service}
}
func (h *AvailabilityHandler) Get(c echo.Context) error {
	startStr := c.QueryParam("start")
	endStr := c.QueryParam("end")
	propertyID := c.QueryParam("property_id")
	roomsStr := c.QueryParam("rooms")
	adultsStr := c.QueryParam("adults")
	childrenStr := c.QueryParam("children")
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")
	if startStr == "" || endStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "start and end dates required"})
	}
	layout := "2006-01-02"
	start, err := time.Parse(layout, startStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid start date format"})
	}
	end, err := time.Parse(layout, endStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid end date format"})
	}
	if !start.Before(end) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "start date must be before end date"})
	}
	rooms := 1
	if roomsStr != "" {
		rooms, _ = strconv.Atoi(roomsStr)
	}
	adults := 1
	if adultsStr != "" {
		adults, _ = strconv.Atoi(adultsStr)
	}
	children := 0
	if childrenStr != "" {
		children, _ = strconv.Atoi(childrenStr)
	}
	page := 1
	limit := 10
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	results, err := h.service.Search(sharedContext.WithRequestID(c.Request().Context(), sharedContext.RequestIDFromEcho(c)), propertyID, start, end, adults, children, rooms)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	totalItems := int64(len(results))
	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit
	var paginatedData []dto.AvailabilityResult
	if startIndex < len(results) {
		if endIndex > len(results) {
			endIndex = len(results)
		}
		paginatedData = results[startIndex:endIndex]
	} else {
		paginatedData = []dto.AvailabilityResult{}
	}
	response := dto.PaginatedResponse[dto.AvailabilityResult]{
		Data: paginatedData,
		Meta: dto.Meta{
			TotalItems: totalItems,
			TotalPages: totalPages,
			Page:       page,
			Limit:      limit,
		},
	}
	return c.JSON(http.StatusOK, response)
}
