package domain

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func RequireSuperAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role, ok := c.Get("role").(string)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		if role != string(RoleSuperAdmin) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "requires super_admin privileges"})
		}
		return next(c)
	}
}
