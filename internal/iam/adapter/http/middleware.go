package http

import (
	"net/http"

	"github.com/ecelayes/pms-backend/internal/iam/domain"
	"github.com/labstack/echo/v4"
)

func RequireRole(allowedRoles ...domain.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roleVal, ok := c.Get("role").(string)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}
			role, err := domain.NewUserRole(roleVal)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}
			for _, allowed := range allowedRoles {
				if role == allowed {
					return next(c)
				}
			}
			return c.JSON(http.StatusForbidden, map[string]string{"error": "insufficient privileges"})
		}
	}
}

func RequireSuperAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return RequireRole(domain.RoleSuperAdmin)(next)
}
