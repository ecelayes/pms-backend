package availability

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/ecelayes/pms-backend/internal/availability/adapter"
	"github.com/ecelayes/pms-backend/internal/availability/adapter/http"
	"github.com/ecelayes/pms-backend/internal/availability/application"
	"github.com/ecelayes/pms-backend/internal/shared/adapter/email"
	redisAdapter "github.com/ecelayes/pms-backend/internal/shared/adapter/redis"
	"github.com/ecelayes/pms-backend/internal/shared/domain"
	catalogApp "github.com/ecelayes/pms-backend/internal/catalog/application"
	pricingApp "github.com/ecelayes/pms-backend/internal/pricing/application"
)

type Module struct {
	Service *application.AvailabilityService
}

func NewModule(
	rdb *redis.Client,
	group *echo.Group,
	catalogService *catalogApp.CatalogService,
	pricingService *pricingApp.PricingService,
	emailSvc *email.Service,
) *Module {
	redisRepo := adapter.NewRedisAvailabilityRepository(rdb)
	svc := application.NewAvailabilityService(redisRepo, catalogService, pricingService)
	eventHandler := application.NewEventHandler(redisRepo)
	analyticsStore := redisAdapter.NewAnalyticsStore(rdb)

	go startStreamConsumers(rdb, eventHandler, emailSvc, analyticsStore)

	httpHandler := http.NewAvailabilityHandler(svc)
	group.GET("/availability", httpHandler.Get)

	return &Module{Service: svc}
}

func startStreamConsumers(client *redis.Client, handler *application.EventHandler, emailSvc *email.Service, analyticsStore *redisAdapter.AnalyticsStore) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		startAvailabilityConsumer(client, handler)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		startNotificationsConsumer(client, emailSvc)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		startAnalyticsConsumer(client, analyticsStore)
	}()

	wg.Wait()
}

func startAvailabilityConsumer(client *redis.Client, handler *application.EventHandler) {
	consumerName := "availability-consumer-1"
	log.Printf("[AvailabilityModule] Starting stream consumer: %s", consumerName)

	consumer := redisAdapter.NewStreamConsumer(client, "availability-group", consumerName)
	ctx := context.Background()

	eventTypes := []string{domain.EventReservationCreated, domain.EventReservationCancelled}
	consumer.Consume(ctx, eventTypes, func(msg *domain.StreamMessage) error {
		return handler.HandleStreamMessage(msg)
	})
}

func startNotificationsConsumer(client *redis.Client, emailSvc *email.Service) {
	consumerName := "notifications-consumer-1"
	log.Printf("[AvailabilityModule] Starting stream consumer: %s", consumerName)

	consumer := redisAdapter.NewStreamConsumer(client, "notifications-group", consumerName)
	ctx := context.Background()

	eventTypes := []string{domain.EventReservationConfirmed, domain.EventReservationCancelled}
	consumer.Consume(ctx, eventTypes, func(msg *domain.StreamMessage) error {
		return handleNotificationEvent(msg, emailSvc)
	})
}

func handleNotificationEvent(msg *domain.StreamMessage, emailSvc *email.Service) error {
	switch msg.EventType {
	case domain.EventReservationConfirmed:
		var payload domain.ReservationConfirmedPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}
		log.Printf("[NotificationsConsumer] Sending confirmation email to %s", payload.GuestEmail)
		return emailSvc.SendReservationConfirmed(payload.GuestEmail, payload.GuestEmail, payload.ReservationCode, time.Now(), time.Now().AddDate(0, 0, 1))

	case domain.EventReservationCancelled:
		var payload domain.ReservationCancelledPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}
		log.Printf("[NotificationsConsumer] Sending cancellation email to %s", payload.GuestEmail)
		return emailSvc.SendReservationCancelled(payload.GuestEmail, payload.GuestEmail, payload.ReservationCode)

	default:
		return nil
	}
}

func startAnalyticsConsumer(client *redis.Client, analyticsStore *redisAdapter.AnalyticsStore) {
	consumerName := "analytics-consumer-1"
	log.Printf("[AvailabilityModule] Starting stream consumer: %s", consumerName)

	consumer := redisAdapter.NewStreamConsumer(client, "analytics-group", consumerName)
	ctx := context.Background()

	eventTypes := []string{domain.EventReservationCreated, domain.EventReservationConfirmed, domain.EventReservationCancelled}
	consumer.Consume(ctx, eventTypes, func(msg *domain.StreamMessage) error {
		return handleAnalyticsEvent(ctx, msg, analyticsStore)
	})
}

func handleAnalyticsEvent(ctx context.Context, msg *domain.StreamMessage, analyticsStore *redisAdapter.AnalyticsStore) error {
	switch msg.EventType {
	case domain.EventReservationCreated:
		var payload domain.ReservationCreatedPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}
		return analyticsStore.RecordEvent(ctx, msg.EventType, payload)

	case domain.EventReservationConfirmed:
		var payload domain.ReservationConfirmedPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}
		return analyticsStore.RecordEvent(ctx, msg.EventType, payload)

	case domain.EventReservationCancelled:
		var payload domain.ReservationCancelledPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}
		return analyticsStore.RecordEvent(ctx, msg.EventType, payload)

	default:
		return nil
	}
}
