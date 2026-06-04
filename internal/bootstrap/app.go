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
	"github.com/ecelayes/pms-backend/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewApp(pool *pgxpool.Pool, rdb *redis.Client) *echo.Echo {
	log, err := logger.New()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	e := echo.New()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogMethod:   true,
		LogLatency:  true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				log.Info("request",
					zap.String("URI", v.URI),
					zap.Int("status", v.Status),
					zap.String("method", v.Method),
					zap.Duration("latency", v.Latency),
				)
			} else {
				log.Error("request error",
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

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("logger", log)
			return next(c)
		}
	})

	e.GET("/health", func(c echo.Context) error {
		ctx := c.Request().Context()
		if err := pool.Ping(ctx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "error", "db": err.Error()})
		}
		if err := rdb.Ping(ctx).Err(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "error", "redis": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	v1 := e.Group("/api/v1")
	protected := v1.Group("")
	admin := e.Group("/admin")

	emailService := email.NewService()

	iamModule := iam.NewModule(pool, v1, protected, emailService)
	catalogModule := catalog.NewModule(pool, v1, protected)
	pricingModule := pricing.NewModule(pool, protected)
	availModule := availability.NewModule(rdb, v1, catalogModule.Service, pricingModule.Service, emailService)
	_ = booking.NewModule(pool, v1, protected, catalogModule.Service, pricingModule.Service, iamModule.UserService, availModule.Service, rdb)
	_ = shared.NewModule(rdb, admin)

	return e
}
