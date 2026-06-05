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
	"log"
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
) *Module {
	repo := adapter.NewPostgresReservationRepository(db)
	var publisher *redisAdapter.StreamProducer
	if redisClient != nil {
		publisher = redisAdapter.NewStreamProducer(redisClient)
		if err := publisher.EnsureGroups(context.Background()); err != nil {
			log.Printf("[BookingModule] Warning: failed to ensure stream groups: %v", err)
		}
	}
	service := application.NewBookingService(
		repo,
		pricingService,
		catalogService,
		userService,
		availService,
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
