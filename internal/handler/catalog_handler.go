package handler

import (
	"net/http"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

type CatalogHandler struct {
	uc *usecase.CatalogUseCase
}

func NewCatalogHandler(uc *usecase.CatalogUseCase) *CatalogHandler {
	return &CatalogHandler{uc: uc}
}

func getRoleFromToken(c echo.Context) string {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims["role"].(string)
}

func (h *CatalogHandler) CreateAmenity(c echo.Context) error {
	role := getRoleFromToken(c)
	
	var req entity.CreateCatalogRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	id, err := h.uc.CreateAmenity(c.Request().Context(), role, req)
	if err != nil {
		if err.Error() == "insufficient permissions" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"id": id})
}

func (h *CatalogHandler) GetAllAmenities(c echo.Context) error {
	list, err := h.uc.GetAllAmenities(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, list)
}

func (h *CatalogHandler) GetAmenityByID(c echo.Context) error {
	id := c.Param("id")

	amenity, err := h.uc.GetAmenityByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "amenity not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, amenity)
}

func (h *CatalogHandler) UpdateAmenity(c echo.Context) error {
	role := getRoleFromToken(c)
	id := c.Param("id")
	var req entity.UpdateCatalogRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	err := h.uc.UpdateAmenity(c.Request().Context(), role, id, req)
	if err != nil {
		if errors.Is(err, entity.ErrInsufficientPermissions) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		}
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "amenity not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "updated successfully"})
}

func (h *CatalogHandler) DeleteAmenity(c echo.Context) error {
	role := getRoleFromToken(c)
	id := c.Param("id")
	
	err := h.uc.DeleteAmenity(c.Request().Context(), role, id)
	if err != nil {
		if errors.Is(err, entity.ErrInsufficientPermissions) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		}
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "amenity not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted successfully"})
}

func (h *CatalogHandler) CreateService(c echo.Context) error {
	role := getRoleFromToken(c)
	var req entity.CreateCatalogRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	id, err := h.uc.CreateService(c.Request().Context(), role, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"id": id})
}

func (h *CatalogHandler) GetAllServices(c echo.Context) error {
	list, err := h.uc.GetAllServices(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, list)
}

func (h *CatalogHandler) GetServiceByID(c echo.Context) error {
	id := c.Param("id")

	service, err := h.uc.GetServiceByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "service not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, service)
}

func (h *CatalogHandler) UpdateService(c echo.Context) error {
	role := getRoleFromToken(c)
	id := c.Param("id")
	var req entity.UpdateCatalogRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	err := h.uc.UpdateService(c.Request().Context(), role, id, req)
	if err != nil {
		if errors.Is(err, entity.ErrInsufficientPermissions) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		}
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "service not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "updated successfully"})
}

func (h *CatalogHandler) DeleteService(c echo.Context) error {
	role := getRoleFromToken(c)
	id := c.Param("id")
	
	err := h.uc.DeleteService(c.Request().Context(), role, id)
	if err != nil {
		if errors.Is(err, entity.ErrInsufficientPermissions) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		}
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "service not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted successfully"})
}
