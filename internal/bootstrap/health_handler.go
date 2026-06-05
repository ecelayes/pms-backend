package bootstrap

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Pinger is the minimal contract for a dependency that can be checked for reachability.
type Pinger interface {
	Ping(ctx context.Context) error
}

// PoolAdapter wraps *pgxpool.Pool to satisfy Pinger.
type PoolAdapter struct {
	PingFn func(ctx context.Context) error
}

func (p *PoolAdapter) Ping(ctx context.Context) error {
	if p.PingFn == nil {
		return nil
	}
	return p.PingFn(ctx)
}

// RedisAdapter wraps *redis.Client to satisfy Pinger.
// go-redis's Ping returns *Status, so we adapt it to error.
type RedisAdapter struct {
	Client *redis.Client
}

func (r *RedisAdapter) Ping(ctx context.Context) error {
	if r.Client == nil {
		return nil
	}
	return r.Client.Ping(ctx).Err()
}

// HealthHandler exposes liveness, readiness and legacy health endpoints.
//
//   - /live  : 200 if the process is alive (no external dependencies checked).
//   - /ready : 200 if all critical dependencies (PostgreSQL, Redis) are reachable.
//   - /health: legacy alias kept for backward compatibility.
type HealthHandler struct {
	pingers map[string]Pinger
	logger  *zap.Logger
}

func newHealthHandler(pingers map[string]Pinger, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{pingers: pingers, logger: logger}
}

// Liveness returns 200 if the process is alive.
// Use this for Kubernetes liveness probes (restart the pod only if this fails).
func (h *HealthHandler) Liveness(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "alive",
	})
}

// Readiness returns 200 only if all critical dependencies are reachable.
// Use this for Kubernetes readiness probes (remove from load balancer if this fails).
func (h *HealthHandler) Readiness(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	checks := map[string]string{}
	httpStatus := http.StatusOK

	for name, p := range h.pingers {
		if err := p.Ping(ctx); err != nil {
			checks[name] = err.Error()
			httpStatus = http.StatusServiceUnavailable
		}
	}

	if httpStatus != http.StatusOK {
		h.logger.Warn("readiness check failed", zap.Any("checks", checks))
		return c.JSON(httpStatus, map[string]any{
			"status": "not_ready",
			"checks": checks,
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "ready",
	})
}

// LegacyHealth preserves the old /health behavior for backward compatibility.
func (h *HealthHandler) LegacyHealth(c echo.Context) error {
	return h.Readiness(c)
}
