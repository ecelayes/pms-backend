package context

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

func TestWithRequestID_AndRequestIDFromContext(t *testing.T) {
	ctx := WithRequestID(context.Background(), "test-id-123")
	got, err := RequestIDFromContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "test-id-123", got)
}

func TestRequestIDFromContext_NoID(t *testing.T) {
	got, err := RequestIDFromContext(context.Background())
	assert.Equal(t, "", got)
	assert.True(t, errors.Is(err, ErrNoRequestID))
}

func TestRequestIDFromContext_NilContext(t *testing.T) {
	got, err := RequestIDFromContext(context.TODO())
	assert.Equal(t, "", got)
	assert.True(t, errors.Is(err, ErrNoRequestID))
}

func TestRequestIDFromContext_EmptyID(t *testing.T) {
	ctx := WithRequestID(context.Background(), "")
	got, err := RequestIDFromContext(ctx)
	assert.Equal(t, "", got)
	assert.True(t, errors.Is(err, ErrNoRequestID))
}

func TestRequestIDFromEcho_ValidContext(t *testing.T) {
	e := echo.New()
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string { return "echo-id" },
		RequestIDHandler: func(c echo.Context, id string) {
			c.Set("request_id", id)
		},
	}))
	e.GET("/test", func(c echo.Context) error {
		rid := RequestIDFromEcho(c)
		return c.String(http.StatusOK, rid)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "echo-id", rec.Body.String())
}

func TestRequestIDFromEcho_NilContext(t *testing.T) {
	rid := RequestIDFromEcho(nil)
	assert.Equal(t, "", rid)
}

func TestRequestIDFromContext_LiteralNil(t *testing.T) {
	var nilCtx context.Context
	got, err := RequestIDFromContext(nilCtx)
	assert.Equal(t, "", got)
	assert.True(t, errors.Is(err, ErrNoRequestID))
}

func TestRequestIDFromEcho_FallbackToHeader(t *testing.T) {
	e := echo.New()
	e.GET("/test", func(c echo.Context) error {
		c.Response().Header().Set("X-Request-ID", "header-id-456")
		rid := RequestIDFromEcho(c)
		return c.String(http.StatusOK, rid)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "header-id-456", rec.Body.String())
}
