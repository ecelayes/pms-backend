package booking

import (
	"context"
	availApp "github.com/ecelayes/pms-backend/internal/availability/application"
	"github.com/ecelayes/pms-backend/internal/booking/adapter"
	"github.com/ecelayes/pms-backend/internal/booking/adapter/http"
	"github.com/ecelayes/pms-backend/internal/booking/application"
	catalogApp "github.com/ecelayes/pms-backend/internal/catalog/application"
	iamApp "github.com/ecelayes/pms-backend/internal/iam/application"
	pricingApp "github.com/ecelayes/pms-backend/internal/pricing/application"
	redisAdapter "github.com/ecelayes/pms-backend/internal/shared/adapter/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Module struct {
	Service *application.BookingService
}

func NewModule(
	db *pgxpool.Pool,
	group *echo.Group,
	protected *echo.Group,
	catalogService *catalogApp.CatalogService,
	pricingService *pricingApp.PricingService,
	userService *iamApp.UserService,
	availService *availApp.AvailabilityService,
	redisClient *redis.Client,
	logger *zap.Logger,
) *Module {
	if logger == nil {
		logger = zap.NewNop()
	}
	repo := adapter.NewPostgresReservationRepository(db)
	var publisher *redisAdapter.StreamProducer
	if redisClient != nil {
		publisher = redisAdapter.NewStreamProducer(redisClient, logger)
		if err := publisher.EnsureGroups(context.Background()); err != nil {
			logger.Warn("failed to ensure stream groups", zap.Error(err))
		}
	}
	service := application.NewBookingService(
		repo,
		pricingService,
		catalogService,
		userService,
		availService,
		logger,
		publisher,
	)
	handler := http.NewReservationHandler(service)
	group.POST("/reservations", handler.Create)
	group.GET("/reservations/:code", handler.GetByCode)
	group.POST("/reservations/:id/cancel", handler.Cancel)
	protected.GET("/reservations/:id/cancel-preview", handler.PreviewCancel)
	protected.DELETE("/reservations/:id", handler.Delete)
	return &Module{Service: service}
}
