package security

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/ecelayes/pms-backend/pkg/auth"
)

type SaltProvider interface {
	GetUserSalt(ctx context.Context, userID string) (string, error)
}

func Auth(provider SaltProvider) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid header format"})
			}
			tokenString := parts[1]

			claims, err := auth.ParseTokenClaimsUnsafe(tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "malformed token"})
			}

			if claims.Purpose != auth.PurposeAuth {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token purpose"})
			}

			userSalt, err := provider.GetUserSalt(c.Request().Context(), claims.UserID)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user not found or inactive"})
			}

			validClaims, err := auth.ValidateSignature(tokenString, userSalt)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token signature"})
			}

			c.Set("user_id", validClaims.UserID)
			c.Set("role", validClaims.Role)

			return next(c)
		}
	}
}
