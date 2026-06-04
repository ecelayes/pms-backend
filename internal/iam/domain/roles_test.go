package domain

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRequireSuperAdminAuthorized(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", string(RoleSuperAdmin))
	
	handler := RequireSuperAdmin(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestRequireSuperAdminForbidden(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", string(RoleUser))
	
	handler := RequireSuperAdmin(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rec.Code)
	}
}

func TestRequireSuperAdminNoRole(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// role not set
	
	handler := RequireSuperAdmin(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestRequireSuperAdminInvalidRoleType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", 12345) // int instead of string
	
	handler := RequireSuperAdmin(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}
