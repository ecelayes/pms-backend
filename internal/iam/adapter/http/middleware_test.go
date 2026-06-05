package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func newMiddlewareContext(roleVal interface{}) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if roleVal != nil {
		c.Set("role", roleVal)
	}
	return c, rec
}

func nextHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func TestRequireSuperAdmin_Authorized(t *testing.T) {
	c, rec := newMiddlewareContext("super_admin")
	handler := RequireSuperAdmin(nextHandler)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRequireSuperAdmin_Forbidden(t *testing.T) {
	c, rec := newMiddlewareContext("user")
	handler := RequireSuperAdmin(nextHandler)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestRequireSuperAdmin_NoRole(t *testing.T) {
	c, rec := newMiddlewareContext(nil)
	handler := RequireSuperAdmin(nextHandler)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireSuperAdmin_InvalidRoleType(t *testing.T) {
	c, rec := newMiddlewareContext(12345) // int instead of string
	handler := RequireSuperAdmin(nextHandler)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireSuperAdmin_InvalidRoleString(t *testing.T) {
	c, rec := newMiddlewareContext("admin") // not a valid role
	handler := RequireSuperAdmin(nextHandler)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "admin" is not a valid role in NewUserRole → returns unauthorized
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireRole_MultipleAllowed(t *testing.T) {
	c, rec := newMiddlewareContext("user")
	handler := RequireRole("user", "super_admin")(nextHandler)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRequireRole_RejectsNonAllowed(t *testing.T) {
	c, rec := newMiddlewareContext("user")
	handler := RequireRole("super_admin")(nextHandler)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}
