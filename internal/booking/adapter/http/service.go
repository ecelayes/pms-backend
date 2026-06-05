package http

import (
	"context"
	"time"

	"github.com/ecelayes/pms-backend/internal/booking/application"
	"github.com/ecelayes/pms-backend/internal/booking/domain"
)

// BookingService defines booking operations exposed to the HTTP layer.
type BookingService interface {
	CreateReservation(
		ctx context.Context,
		unitTypeID, ratePlanID string,
		start, end time.Time,
		guestEmail, guestFirstName, guestLastName, guestPhone string,
		adults, children int,
	) (string, error)
	GetReservationByCode(ctx context.Context, code string) (*domain.Reservation, error)
	CancelReservation(ctx context.Context, id string) error
	PreviewCancellation(ctx context.Context, id string) (float64, error)
}

var _ BookingService = (*application.BookingService)(nil)
