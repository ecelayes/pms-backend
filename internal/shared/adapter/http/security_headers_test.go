package http

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestSecurityHeaders_DefaultHeaders(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := SecurityHeaders()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"X-XSS-Protection":          "0",
		"Cross-Origin-Opener-Policy": "same-origin",
	}
	for k, want := range expected {
		if got := rec.Header().Get(k); got != want {
			t.Errorf("header %s: expected %q, got %q", k, want, got)
		}
	}
}

func TestSecurityHeaders_HSTSOnlyOnTLS(t *testing.T) {
	tests := []struct {
		name   string
		isTLS  bool
		expect string
	}{
		{"TLS connection sets HSTS", true, "max-age=31536000; includeSubDomains"},
		{"plain HTTP omits HSTS", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.isTLS {
				req.TLS = &tls.ConnectionState{}
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := SecurityHeaders()(func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			})
			if err := handler(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := rec.Header().Get("Strict-Transport-Security")
			if got != tt.expect {
				t.Errorf("HSTS: expected %q, got %q", tt.expect, got)
			}
		})
	}
}

func TestSecurityHeaders_ContentSecurityPolicyDefault(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := SecurityHeaders()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	csp := rec.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Fatal("CSP header is empty")
	}
	// Must lock down script/frame sources
	if csp != "default-src 'none'; frame-ancestors 'none'" {
		t.Errorf("unexpected CSP: %q", csp)
	}
}

func TestSecurityHeaders_CSPAllowCustom(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := SecurityHeadersWithConfig(SecurityHeadersConfig{
		ContentSecurityPolicy: "default-src 'self'",
	})(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := rec.Header().Get("Content-Security-Policy"); got != "default-src 'self'" {
		t.Errorf("CSP: expected %q, got %q", "default-src 'self'", got)
	}
}

func TestSecurityHeaders_DoesNotOverrideExisting(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	rec.Header().Set("X-Frame-Options", "SAMEORIGIN")
	c := e.NewContext(req, rec)

	handler := SecurityHeaders()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// pre-existing headers MUST NOT be overridden
	if got := rec.Header().Get("X-Frame-Options"); got != "SAMEORIGIN" {
		t.Errorf("pre-existing X-Frame-Options was overridden: got %q", got)
	}
}
