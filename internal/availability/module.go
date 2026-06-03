package availability

import (
	"context"
	"github.com/ecelayes/pms-backend/internal/availability/adapter"
	"github.com/ecelayes/pms-backend/internal/availability/adapter/http"
	"github.com/ecelayes/pms-backend/internal/availability/application"
	catalogApp "github.com/ecelayes/pms-backend/internal/catalog/application"
	pricingApp "github.com/ecelayes/pms-backend/internal/pricing/application"
	"github.com/ecelayes/pms-backend/internal/shared/domain"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"log"
	"sync"
)

type Module struct {
	Service *application.AvailabilityService
}

func NewModule(
	rdb *redis.Client,
	group *echo.Group,
	catalogService *catalogApp.CatalogService,
	pricingService *pricingApp.PricingService,
) *Module {
	redisRepo := adapter.NewRedisAvailabilityRepository(rdb)
	svc := application.NewAvailabilityService(redisRepo, catalogService, pricingService)
	eventHandler := application.NewEventHandler(redisRepo)
	go startStreamConsumers(rdb, eventHandler)
	httpHandler := http.NewAvailabilityHandler(svc)
	group.GET("/search", httpHandler.Get)
	return &Module{
		Service: svc,
	}
}
func startStreamConsumers(client *redis.Client, handler *application.EventHandler) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		startAvailabilityConsumer(client, handler)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		startNotificationsConsumer(client)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		startAnalyticsConsumer(client)
	}()
	wg.Wait()
}
func startAvailabilityConsumer(client *redis.Client, handler *application.EventHandler) {
	consumerName := "availability-consumer-1"
	log.Printf("[AvailabilityModule] Starting stream consumer: %s", consumerName)
	consumer := domain.NewRedisStreamConsumer(client, domain.GroupAvailability, consumerName)
	ctx := context.Background()
	eventTypes := []string{domain.EventReservationCreated, domain.EventReservationCancelled}
	consumer.Consume(ctx, eventTypes, func(msg *domain.StreamMessage) error {
		return handler.HandleStreamMessage(msg)
	})
}
func startNotificationsConsumer(client *redis.Client) {
	consumerName := "notifications-consumer-1"
	log.Printf("[AvailabilityModule] Starting stream consumer: %s (placeholder)", consumerName)
	consumer := domain.NewRedisStreamConsumer(client, domain.GroupNotifications, consumerName)
	ctx := context.Background()
	eventTypes := []string{domain.EventReservationConfirmed, domain.EventReservationCancelled}
	consumer.Consume(ctx, eventTypes, func(msg *domain.StreamMessage) error {
		log.Printf("[NotificationsConsumer] Would send notification for event %s (placeholder)", msg.EventType)
		return nil
	})
}
func startAnalyticsConsumer(client *redis.Client) {
	consumerName := "analytics-consumer-1"
	log.Printf("[AvailabilityModule] Starting stream consumer: %s (placeholder)", consumerName)
	consumer := domain.NewRedisStreamConsumer(client, domain.GroupAnalytics, consumerName)
	ctx := context.Background()
	eventTypes := []string{domain.EventReservationCreated, domain.EventReservationConfirmed, domain.EventReservationCancelled}
	consumer.Consume(ctx, eventTypes, func(msg *domain.StreamMessage) error {
		log.Printf("[AnalyticsConsumer] Would record analytics for event %s (placeholder)", msg.EventType)
		return nil
	})
}
