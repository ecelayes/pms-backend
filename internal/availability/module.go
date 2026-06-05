package availability

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

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
	logger *zap.Logger,
) *Module {
	if logger == nil {
		logger = zap.NewNop()
	}
	redisRepo := adapter.NewRedisAvailabilityRepository(rdb)
	svc := application.NewAvailabilityService(redisRepo, catalogService, pricingService)
	eventHandler := application.NewEventHandler(redisRepo, logger)
	analyticsStore := redisAdapter.NewAnalyticsStore(rdb, logger)

	go startStreamConsumers(rdb, eventHandler, emailSvc, analyticsStore, logger)

	httpHandler := http.NewAvailabilityHandler(svc)
	group.GET("/availability", httpHandler.Get)

	return &Module{Service: svc}
}

func startStreamConsumers(client *redis.Client, handler *application.EventHandler, emailSvc *email.Service, analyticsStore *redisAdapter.AnalyticsStore, logger *zap.Logger) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		startAvailabilityConsumer(client, handler, logger)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		startNotificationsConsumer(client, emailSvc, logger)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		startAnalyticsConsumer(client, analyticsStore, logger)
	}()

	wg.Wait()
}

func startAvailabilityConsumer(client *redis.Client, handler *application.EventHandler, logger *zap.Logger) {
	consumerName := "availability-consumer-1"
	logger.Info("starting stream consumer", zap.String("consumer", consumerName))

	consumer := redisAdapter.NewStreamConsumer(client, "availability-group", consumerName, logger)
	ctx := context.Background()

	eventTypes := []string{domain.EventReservationCreated, domain.EventReservationCancelled}
	consumer.Consume(ctx, eventTypes, func(msg *domain.StreamMessage) error {
		return handler.HandleStreamMessage(msg)
	})
}

func startNotificationsConsumer(client *redis.Client, emailSvc *email.Service, logger *zap.Logger) {
	consumerName := "notifications-consumer-1"
	logger.Info("starting stream consumer", zap.String("consumer", consumerName))

	consumer := redisAdapter.NewStreamConsumer(client, "notifications-group", consumerName, logger)
	ctx := context.Background()

	eventTypes := []string{domain.EventReservationConfirmed, domain.EventReservationCancelled}
	consumer.Consume(ctx, eventTypes, func(msg *domain.StreamMessage) error {
		return handleNotificationEvent(msg, emailSvc, logger)
	})
}

func handleNotificationEvent(msg *domain.StreamMessage, emailSvc *email.Service, logger *zap.Logger) error {
	switch msg.EventType {
	case domain.EventReservationConfirmed:
		var payload domain.ReservationConfirmedPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}
		logger.Info("sending confirmation email", zap.String("guest_email", payload.GuestEmail))
		return emailSvc.SendReservationConfirmed(payload.GuestEmail, payload.GuestEmail, payload.ReservationCode, time.Now(), time.Now().AddDate(0, 0, 1))

	case domain.EventReservationCancelled:
		var payload domain.ReservationCancelledPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}
		logger.Info("sending cancellation email", zap.String("guest_email", payload.GuestEmail))
		return emailSvc.SendReservationCancelled(payload.GuestEmail, payload.GuestEmail, payload.ReservationCode)

	default:
		return nil
	}
}

func startAnalyticsConsumer(client *redis.Client, analyticsStore *redisAdapter.AnalyticsStore, logger *zap.Logger) {
	consumerName := "analytics-consumer-1"
	logger.Info("starting stream consumer", zap.String("consumer", consumerName))

	consumer := redisAdapter.NewStreamConsumer(client, "analytics-group", consumerName, logger)
	ctx := context.Background()

	eventTypes := []string{domain.EventReservationCreated, domain.EventReservationConfirmed, domain.EventReservationCancelled}
	consumer.Consume(ctx, eventTypes, func(msg *domain.StreamMessage) error {
		return handleAnalyticsEvent(ctx, msg, analyticsStore, logger)
	})
}

func handleAnalyticsEvent(ctx context.Context, msg *domain.StreamMessage, analyticsStore *redisAdapter.AnalyticsStore, logger *zap.Logger) error {
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
