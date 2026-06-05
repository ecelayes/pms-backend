package http

import (
	"errors"
	"github.com/ecelayes/pms-backend/internal/catalog/application"
	"github.com/ecelayes/pms-backend/internal/catalog/domain"
	"github.com/ecelayes/pms-backend/internal/shared/dto"
	"github.com/labstack/echo/v4"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type CatalogHandler struct {
	service CatalogService
}

func NewCatalogHandler(service CatalogService) *CatalogHandler {
	return &CatalogHandler{service: service}
}

type CreateAmenityRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}
type CreateGuestServiceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}
type UpdateCatalogItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

func (h *CatalogHandler) CreateAmenity(c echo.Context) error {
	var req CreateAmenityRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	id, err := h.service.CreateAmenity(c.Request().Context(), req.Name, req.Description, req.Icon)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.JSON(http.StatusConflict, map[string]string{"error": "amenity already exists"})
		}
		if strings.Contains(err.Error(), "invalid") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"id": id})
}
func (h *CatalogHandler) GetAllAmenities(c echo.Context) error {
	page, limit := 1, 10
	if p, err := strconv.Atoi(c.QueryParam("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(c.QueryParam("limit")); err == nil && l > 0 {
		limit = l
	}
	list, totalCount, err := h.service.ListAmenities(c.Request().Context(), page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	dtos := make([]dto.Amenity, len(list))
	for i, item := range list {
		dtos[i] = h.toAmenityDTO(item)
	}
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
	response := dto.PaginatedResponse[dto.Amenity]{
		Data: dtos,
		Meta: dto.Meta{
			TotalItems: totalCount,
			TotalPages: totalPages,
			Page:       page,
			Limit:      limit,
		},
	}
	return c.JSON(http.StatusOK, response)
}
func (h *CatalogHandler) GetAmenityByID(c echo.Context) error {
	id := c.Param("id")
	amenity, err := h.service.GetAmenity(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "amenity not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.toAmenityDTO(amenity))
}
func (h *CatalogHandler) UpdateAmenity(c echo.Context) error {
	id := c.Param("id")
	var req UpdateCatalogItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	err := h.service.UpdateAmenity(c.Request().Context(), id, req.Name, req.Description, req.Icon)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "amenity not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "updated successfully"})
}
func (h *CatalogHandler) DeleteAmenity(c echo.Context) error {
	id := c.Param("id")
	err := h.service.DeleteAmenity(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted successfully"})
}
func (h *CatalogHandler) CreateService(c echo.Context) error {
	var req CreateGuestServiceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	id, err := h.service.CreateGuestService(c.Request().Context(), req.Name, req.Description, req.Icon)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.JSON(http.StatusConflict, map[string]string{"error": "service already exists"})
		}
		if strings.Contains(err.Error(), "invalid") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"id": id})
}
func (h *CatalogHandler) GetAllServices(c echo.Context) error {
	page, limit := 1, 10
	if p, err := strconv.Atoi(c.QueryParam("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(c.QueryParam("limit")); err == nil && l > 0 {
		limit = l
	}
	list, totalCount, err := h.service.ListGuestServices(c.Request().Context(), page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	dtos := make([]dto.GuestService, len(list))
	for i, item := range list {
		dtos[i] = h.toGuestServiceDTO(item)
	}
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
	response := dto.PaginatedResponse[dto.GuestService]{
		Data: dtos,
		Meta: dto.Meta{
			TotalItems: totalCount,
			TotalPages: totalPages,
			Page:       page,
			Limit:      limit,
		},
	}
	return c.JSON(http.StatusOK, response)
}
func (h *CatalogHandler) GetServiceByID(c echo.Context) error {
	id := c.Param("id")
	service, err := h.service.GetGuestService(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "service not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.toGuestServiceDTO(service))
}
func (h *CatalogHandler) UpdateService(c echo.Context) error {
	id := c.Param("id")
	var req UpdateCatalogItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	err := h.service.UpdateGuestService(c.Request().Context(), id, req.Name, req.Description, req.Icon)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "service not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "updated successfully"})
}
func (h *CatalogHandler) DeleteService(c echo.Context) error {
	id := c.Param("id")
	err := h.service.DeleteGuestService(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted successfully"})
}
func (h *CatalogHandler) toAmenityDTO(a *domain.Amenity) dto.Amenity {
	return dto.Amenity{
		ID:          a.ID(),
		Name:        a.Name(),
		Description: a.Description(),
		Icon:        a.Icon(),
	}
}
func (h *CatalogHandler) toGuestServiceDTO(s *domain.GuestService) dto.GuestService {
	return dto.GuestService{
		ID:          s.ID(),
		Name:        s.Name(),
		Description: s.Description(),
		Icon:        s.Icon(),
	}
}
