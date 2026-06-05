package bootstrap

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"

	sharedHTTP "github.com/ecelayes/pms-backend/internal/shared/adapter/http"
)

func TestRequestID_GeneratesID(t *testing.T) {
	e := echo.New()
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return "test-id-123"
		},
		RequestIDHandler: func(c echo.Context, id string) {
			c.Set("request_id", id)
		},
	}))

	e.GET("/test", func(c echo.Context) error {
		rid, _ := c.Get("request_id").(string)
		return c.String(http.StatusOK, rid)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test-id-123", rec.Header().Get(echo.HeaderXRequestID))
	assert.Equal(t, "test-id-123", rec.Body.String())
}

func TestRequestID_PreservesIncomingID(t *testing.T) {
	e := echo.New()
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return "generated-id"
		},
	}))

	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(echo.HeaderXRequestID, "incoming-id-456")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "incoming-id-456", rec.Header().Get(echo.HeaderXRequestID))
}

func TestRequestID_LogsIncludeRequestID(t *testing.T) {
	e := echo.New()
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return "log-test-id"
		},
	}))

	var loggedID string
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogRequestID: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			loggedID = v.RequestID
			return nil
		},
	}))

	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, "log-test-id", loggedID)
	assert.Equal(t, "log-test-id", rec.Header().Get(echo.HeaderXRequestID))
}

func TestLoggerMiddleware_SetsContextValues(t *testing.T) {
	e := echo.New()
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return "ctx-id"
		},
		RequestIDHandler: func(c echo.Context, id string) {
			c.Set("request_id", id)
		},
	}))

	var capturedLogger interface{}
	var capturedRequestID string
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("logger", "mock-logger")
			return next(c)
		}
	})

	e.GET("/test", func(c echo.Context) error {
		capturedLogger = c.Get("logger")
		capturedRequestID, _ = c.Get("request_id").(string)
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.NotNil(t, capturedLogger)
	assert.Equal(t, "mock-logger", capturedLogger)
	assert.Equal(t, "ctx-id", capturedRequestID)
}

func TestMetrics_EndpointExposed(t *testing.T) {
	e := echo.New()
	e.Use(sharedHTTP.Metrics())
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	e.GET("/ping", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)

	req2 := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req2)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "pms_http_requests_total")
	assert.Contains(t, body, "pms_http_request_duration_seconds")
	assert.Contains(t, body, "pms_http_requests_in_flight")
}
