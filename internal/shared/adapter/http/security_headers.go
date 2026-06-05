package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	defaultContentSecurityPolicy = "default-src 'none'; frame-ancestors 'none'"
	defaultStrictTransportSecurity = "max-age=31536000; includeSubDomains"
)

type SecurityHeadersConfig struct {
	ContentSecurityPolicy string
	StrictTransportSecurity string
}

func SecurityHeaders() echo.MiddlewareFunc {
	return SecurityHeadersWithConfig(SecurityHeadersConfig{})
}

func SecurityHeadersWithConfig(cfg SecurityHeadersConfig) echo.MiddlewareFunc {
	csp := cfg.ContentSecurityPolicy
	if csp == "" {
		csp = defaultContentSecurityPolicy
	}
	hsts := cfg.StrictTransportSecurity
	if hsts == "" {
		hsts = defaultStrictTransportSecurity
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Response().Header()
			setIfMissing(h, "X-Content-Type-Options", "nosniff")
			setIfMissing(h, "X-Frame-Options", "DENY")
			setIfMissing(h, "Referrer-Policy", "strict-origin-when-cross-origin")
			setIfMissing(h, "X-XSS-Protection", "0")
			setIfMissing(h, "Cross-Origin-Opener-Policy", "same-origin")
			setIfMissing(h, "Content-Security-Policy", csp)
			if c.Request().TLS != nil {
				setIfMissing(h, "Strict-Transport-Security", hsts)
			}
			return next(c)
		}
	}
}

func setIfMissing(h http.Header, key, value string) {
	if h.Get(key) == "" {
		h.Set(key, value)
	}
}
