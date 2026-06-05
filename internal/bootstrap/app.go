package bootstrap

import (
	"net/http"
	"os"
	"strings"

	"github.com/ecelayes/pms-backend/internal/availability"
	"github.com/ecelayes/pms-backend/internal/booking"
	"github.com/ecelayes/pms-backend/internal/catalog"
	"github.com/ecelayes/pms-backend/internal/iam"
	"github.com/ecelayes/pms-backend/internal/pricing"
	"github.com/ecelayes/pms-backend/internal/shared"
	"github.com/ecelayes/pms-backend/internal/shared/adapter/email"
	sharedHTTP "github.com/ecelayes/pms-backend/internal/shared/adapter/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/google/uuid"
)

func NewApp(pool *pgxpool.Pool, rdb *redis.Client, appLogger *zap.Logger) *echo.Echo {
	log := appLogger

	e := echo.New()

	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: uuid.NewString,
		RequestIDHandler: func(c echo.Context, id string) {
			c.Set("request_id", id)
		},
	}))

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:        true,
		LogStatus:     true,
		LogMethod:     true,
		LogLatency:    true,
		LogError:      true,
		LogRequestID:  true,
		HandleError:   true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				log.Info("request",
					zap.String("request_id", v.RequestID),
					zap.String("URI", v.URI),
					zap.Int("status", v.Status),
					zap.String("method", v.Method),
					zap.Duration("latency", v.Latency),
				)
			} else {
				log.Error("request error",
					zap.String("request_id", v.RequestID),
					zap.String("URI", v.URI),
					zap.Int("status", v.Status),
					zap.String("method", v.Method),
					zap.Duration("latency", v.Latency),
					zap.Error(v.Error),
				)
			}
			return nil
		},
	}))

	e.Use(middleware.Recover())

	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "*"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: strings.Split(allowedOrigins, ","),
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))

	e.Use(sharedHTTP.SecurityHeaders())
	e.Use(sharedHTTP.Metrics())

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("logger", log)
			if rid, ok := c.Get("request_id").(string); !ok || rid == "" {
				c.Set("request_id", c.Response().Header().Get(echo.HeaderXRequestID))
			}
			return next(c)
		}
	})

	pingers := map[string]Pinger{
		"db":    &PoolAdapter{PingFn: pool.Ping},
		"redis": &RedisAdapter{Client: rdb},
	}
	healthHandler := newHealthHandler(pingers, log)

	e.GET("/live", healthHandler.Liveness)
	e.GET("/ready", healthHandler.Readiness)
	e.GET("/health", healthHandler.LegacyHealth)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	v1 := e.Group("/api/v1")
	protected := v1.Group("")
	admin := e.Group("/admin")

	emailService := email.NewService()

	iamModule := iam.NewModule(pool, v1, protected, emailService)
	catalogModule := catalog.NewModule(pool, v1, protected)
	pricingModule := pricing.NewModule(pool, protected)
	availModule := availability.NewModule(rdb, v1, catalogModule.Service, pricingModule.Service, emailService, log)
	_ = booking.NewModule(pool, v1, protected, catalogModule.Service, pricingModule.Service, iamModule.UserService, availModule.Service, rdb, log)
	_ = shared.NewModule(rdb, admin, log)

	return e
}
