package token

import (
	"context"
	"github.com/ecelayes/pms-backend/pkg/auth"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

type SaltProvider interface {
	GetUserSalt(ctx context.Context, userID string) (string, error)
}

// AuthConfig bundles the dependencies for the Auth middleware.
//
// SECURITY: tokenGenerator is the only crypto component used here.
// The production wiring passes a JWTTokenGenerator (with alg confusion protection).
type AuthConfig struct {
	SaltProvider   SaltProvider
	TokenGenerator auth.TokenGenerator
}

// Auth returns an Echo middleware that validates JWT tokens.
//
// SECURITY: The token generator is injected, so the actual signing algorithm
// is determined at startup time and cannot be changed at runtime.
func Auth(cfg AuthConfig) echo.MiddlewareFunc {
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

			claims, err := cfg.TokenGenerator.ParseUnsafe(tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "malformed token"})
			}
			if claims.Purpose != auth.PurposeAuth {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token purpose"})
			}
			userSalt, err := cfg.SaltProvider.GetUserSalt(c.Request().Context(), claims.UserID)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user not found or inactive"})
			}
			validClaims, err := cfg.TokenGenerator.VerifySignature(tokenString, userSalt)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token signature"})
			}
			c.Set("user_id", validClaims.UserID)
			c.Set("organization_id", validClaims.OrganizationID)
			c.Set("role", validClaims.Role)
			return next(c)
		}
	}
}
