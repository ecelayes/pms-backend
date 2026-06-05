package application

import (
	"context"
	"errors"
	"github.com/ecelayes/pms-backend/internal/booking/domain"
	sharedDomain "github.com/ecelayes/pms-backend/internal/shared/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"log"
	"time"

	"go.uber.org/zap"
)

var (
	ErrNoAvailability = errors.New("no availability for selected dates")
)

type PricingService interface {
	CalculateBasePrice(ctx context.Context, unitTypeID string, defaultPrice vo.Money, start, end time.Time) (vo.Money, error)
	CalculateStayPrice(ctx context.Context, unitTypeID, ratePlanID string, defaultPrice vo.Money, start, end time.Time, adults, children int) (vo.Money, error)
	CalculateCancellationPenalty(ctx context.Context, ratePlanID string, reservationStart time.Time, cancelDate time.Time, totalPrice vo.Money) (vo.Money, error)
}
type CatalogService interface {
	GetUnitType(ctx context.Context, id string) (propertyID string, basePrice vo.Money, totalQty int, err error)
}
type IdentityService interface {
	FindOrCreateGuest(ctx context.Context, email, firstName, lastName, phone string) (guestID string, err error)
}
type AvailabilityService interface {
	UpdateInventory(ctx context.Context, propertyID, unitTypeID string, start, end time.Time, delta int) error
}

// ReservationFactory builds domain.Reservation aggregates from external inputs.
// Default implementation calls domain.NewReservation. Tests can inject mocks
// that return pre-built reservations to exercise error paths (e.g. Confirm on
// an already-confirmed reservation).
type ReservationFactory interface {
	Create(
		propertyID, unitTypeID, ratePlanID, guestID string,
		dateRange vo.DateRange,
		price vo.Money,
		guestEmail string,
	) (*domain.Reservation, error)
}
// StreamPublisher is the contract for publishing reservation events.
// We keep this as a focused interface (only 2 methods) because the booking
// service only needs these. The full Redis adapter satisfies it via Go's
// structural typing (no explicit "implements" needed).
type StreamPublisher interface {
	PublishReservationCreated(ctx context.Context, payload sharedDomain.ReservationCreatedPayload) error
	PublishReservationCancelled(ctx context.Context, payload sharedDomain.ReservationCancelledPayload) error
}
type BookingEventPublisher = StreamPublisher

// defaultReservationFactory is the production implementation of ReservationFactory.
// It calls domain.NewReservation to build a fresh aggregate in pending status.
type defaultReservationFactory struct{}

func (f *defaultReservationFactory) Create(
	propertyID, unitTypeID, ratePlanID, guestID string,
	dateRange vo.DateRange,
	price vo.Money,
	guestEmail string,
) (*domain.Reservation, error) {
	return domain.NewReservation(propertyID, unitTypeID, ratePlanID, guestID, dateRange, price, guestEmail)
}
type BookingService struct {
	repo         domain.ReservationRepository
	pricing      PricingService
	catalog      CatalogService
	identity     IdentityService
	availability AvailabilityService
	publisher    StreamPublisher
	factory      ReservationFactory
	logger       *zap.Logger
}

func NewBookingService(
	repo domain.ReservationRepository,
	pricing PricingService,
	catalog CatalogService,
	identity IdentityService,
	availability AvailabilityService,
	logger *zap.Logger,
	publishers ...StreamPublisher,
) *BookingService {
	if logger == nil {
		logger = zap.NewNop()
	}
	var publisher StreamPublisher
	if len(publishers) > 0 && publishers[0] != nil {
		publisher = publishers[0]
	}
	return &BookingService{
		repo:         repo,
		pricing:      pricing,
		catalog:      catalog,
		identity:     identity,
		availability: availability,
		publisher:    publisher,
		factory:      &defaultReservationFactory{},
		logger:       logger,
	}
}
func (s *BookingService) CreateReservation(
	ctx context.Context,
	unitTypeID string,
	ratePlanID string,
	start, end time.Time,
	guestEmail, guestFirstName, guestLastName, guestPhone string,
	adults, children int,
) (string, error) {
	dr, err := vo.NewDateRange(start, end)
	if err != nil {
		return "", err
	}
	guestID, err := s.identity.FindOrCreateGuest(ctx, guestEmail, guestFirstName, guestLastName, guestPhone)
	if err != nil {
		return "", err
	}
	var reservationCode string
	var propertyID string
	err = s.repo.RunInTransaction(ctx, func(ctx context.Context) error {
		if err := s.repo.LockUnitType(ctx, unitTypeID); err != nil {
			return err
		}
		var basePrice vo.Money
		var totalQty int
		propertyID, basePrice, totalQty, err = s.catalog.GetUnitType(ctx, unitTypeID)
		if err != nil {
			return err
		}
		count, err := s.repo.CountOverlapping(ctx, unitTypeID, start, end)
		if err != nil {
			return err
		}
		if (totalQty - count) < 1 {
			return ErrNoAvailability
		}
		price, err := s.pricing.CalculateStayPrice(ctx, unitTypeID, ratePlanID, basePrice, start, end, adults, children)
		if err != nil {
			return err
		}
		res, err := s.factory.Create(propertyID, unitTypeID, ratePlanID, guestID, dr, price, guestEmail)
		if err != nil {
			return err
		}
		if err := res.Confirm(); err != nil {
			return err
		}
		if err := s.repo.Save(ctx, res); err != nil {
			return err
		}
		reservationCode = res.ReservationCode()
		return nil
	})
	if err != nil {
		return "", err
	}
	go func() {
		if s.publisher != nil {
			res, err := s.repo.GetByCode(ctx, reservationCode)
			if err != nil || res == nil {
				log.Printf("[BookingService] Failed to get reservation for event: %v", err)
				return
			}
			event := domain.NewReservationCreatedEvent(res)
			payload := event.ToPayload()
			if pubErr := s.publisher.PublishReservationCreated(ctx, payload); pubErr != nil {
				log.Printf("[BookingService] Failed to publish reservation.created event: %v", pubErr)
			} else {
				log.Printf("[BookingService] Published reservation.created event to stream for %s", reservationCode)
			}
		}
	}()
	return reservationCode, nil
}
func (s *BookingService) GetReservationByCode(ctx context.Context, code string) (*domain.Reservation, error) {
	return s.repo.GetByCode(ctx, code)
}
func (s *BookingService) CancelReservation(ctx context.Context, id string) error {
	res, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("reservation not found")
	}
	res.Cancel()
	if err := s.repo.Update(ctx, res); err != nil {
		return err
	}
	go func() {
		if s.publisher != nil {
			event := domain.NewReservationCancelledEvent(res)
			payload := event.ToPayload()
			if pubErr := s.publisher.PublishReservationCancelled(ctx, payload); pubErr != nil {
				log.Printf("[BookingService] Failed to publish reservation.cancelled event: %v", pubErr)
			} else {
				log.Printf("[BookingService] Published reservation.cancelled event to stream for %s", res.ReservationCode())
			}
		}
	}()
	return nil
}
func (s *BookingService) PreviewCancellation(ctx context.Context, id string) (float64, error) {
	res, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return 0, err
	}
	if res == nil {
		return 0, errors.New("reservation not found")
	}
	if res.Status() == domain.StatusCancelled {
		return 0, errors.New("reservation already cancelled")
	}
	cancelDate := time.Now()
	penalty, err := s.pricing.CalculateCancellationPenalty(
		ctx,
		res.RatePlanID(),
		res.DateRange().Start(),
		cancelDate,
		res.Price(),
	)
	if err != nil {
		return 0, err
	}
	return float64(penalty.Amount()) / 100.0, nil
}
