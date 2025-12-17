package handler

import (
	"net/http"
	"strconv"
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

type OrganizationHandler struct {
	uc *usecase.OrganizationUseCase
}

func NewOrganizationHandler(uc *usecase.OrganizationUseCase) *OrganizationHandler {
	return &OrganizationHandler{uc: uc}
}

func (h *OrganizationHandler) Create(c echo.Context) error {
	var req entity.CreateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	id, err := h.uc.Create(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"organization_id": id})
}

func (h *OrganizationHandler) GetAll(c echo.Context) error {
	pageParam := c.QueryParam("page")
	limitParam := c.QueryParam("limit")

	var pagination entity.PaginationRequest

	if pageParam == "" && limitParam == "" {
		pagination = entity.PaginationRequest{Unlimited: true}
	} else {
		page, limit := 1, 10
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 { page = p }
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 { limit = l }
		pagination = entity.PaginationRequest{Page: page, Limit: limit}
	}

	orgs, total, err := h.uc.GetAll(c.Request().Context(), pagination)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	var totalPages int
	var limit int
	var currentPage int

	if pagination.Unlimited {
		totalPages = 1
		limit = int(total)
		currentPage = 1
	} else {
		limit = pagination.Limit
		currentPage = pagination.Page
		totalPages = int(total) / limit
		if int(total)%limit != 0 {
			totalPages++
		}
	}

	response := entity.PaginatedResponse[entity.Organization]{
		Data: orgs,
		Meta: entity.PaginationMeta{
			Page:       currentPage,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}
	return c.JSON(http.StatusOK, response)
}

func (h *OrganizationHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	org, err := h.uc.GetByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, org)
}

func (h *OrganizationHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req entity.UpdateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	if err := h.uc.Update(c.Request().Context(), id, req); err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		if errors.Is(err, entity.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "updated successfully"})
}

func (h *OrganizationHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.uc.Delete(c.Request().Context(), id); err != nil {
		if errors.Is(err, entity.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted successfully"})
}
