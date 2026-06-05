package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// mockPinger is a stub Pinger for testing.
type mockPinger struct {
	err error
}

func (m *mockPinger) Ping(_ context.Context) error {
	return m.err
}

func TestLiveness_AlwaysReturns200(t *testing.T) {
	e := echo.New()
	h := newHealthHandler(map[string]Pinger{}, zap.NewNop())
	e.GET("/live", h.Liveness)

	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var body map[string]string
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "alive", body["status"])
}

func TestReadiness_AllPingersOK_Returns200(t *testing.T) {
	e := echo.New()
	pingers := map[string]Pinger{
		"db":    &mockPinger{},
		"redis": &mockPinger{},
	}
	h := newHealthHandler(pingers, zap.NewNop())
	e.GET("/ready", h.Readiness)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var body map[string]string
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "ready", body["status"])
}

func TestReadiness_DBPingerFails_Returns503(t *testing.T) {
	e := echo.New()
	pingers := map[string]Pinger{
		"db":    &mockPinger{err: errors.New("db down")},
		"redis": &mockPinger{},
	}
	h := newHealthHandler(pingers, zap.NewNop())
	e.GET("/ready", h.Readiness)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "not_ready", body["status"])
	checks, ok := body["checks"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "db down", checks["db"])
	assert.NotContains(t, checks, "redis")
}

func TestReadiness_RedisPingerFails_Returns503(t *testing.T) {
	e := echo.New()
	pingers := map[string]Pinger{
		"db":    &mockPinger{},
		"redis": &mockPinger{err: errors.New("redis down")},
	}
	h := newHealthHandler(pingers, zap.NewNop())
	e.GET("/ready", h.Readiness)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "not_ready", body["status"])
	checks, ok := body["checks"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "redis down", checks["redis"])
}

func TestReadiness_AllPingersFail_ReturnsAllErrors(t *testing.T) {
	e := echo.New()
	pingers := map[string]Pinger{
		"db":    &mockPinger{err: errors.New("db down")},
		"redis": &mockPinger{err: errors.New("redis down")},
	}
	h := newHealthHandler(pingers, zap.NewNop())
	e.GET("/ready", h.Readiness)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	checks, ok := body["checks"].(map[string]any)
	assert.True(t, ok)
	assert.Len(t, checks, 2)
}

func TestLegacyHealth_DelegatesToReadiness(t *testing.T) {
	e := echo.New()
	pingers := map[string]Pinger{
		"db": &mockPinger{err: errors.New("down")},
	}
	h := newHealthHandler(pingers, zap.NewNop())
	e.GET("/health", h.LegacyHealth)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestPoolAdapter_PingUsesProvidedFn(t *testing.T) {
	called := false
	adapter := &PoolAdapter{
		PingFn: func(_ context.Context) error {
			called = true
			return nil
		},
	}
	err := adapter.Ping(context.Background())
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestPoolAdapter_NilPingFn_ReturnsNil(t *testing.T) {
	adapter := &PoolAdapter{}
	assert.NoError(t, adapter.Ping(context.Background()))
}

func TestPoolAdapter_PingFnError_PropagatesError(t *testing.T) {
	adapter := &PoolAdapter{
		PingFn: func(_ context.Context) error {
			return errors.New("ping failed")
		},
	}
	assert.Error(t, adapter.Ping(context.Background()))
}
