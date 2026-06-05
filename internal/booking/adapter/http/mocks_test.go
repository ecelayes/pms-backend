package http

import (
	"context"
	"time"

	"github.com/ecelayes/pms-backend/internal/booking/domain"
)

type mockBookingService struct {
	createResult string
	createErr    error
	getResult    *domain.Reservation
	getErr       error
	cancelErr    error
	previewAmt   float64
	previewErr   error
}

func (m *mockBookingService) CreateReservation(
	ctx context.Context,
	unitTypeID, ratePlanID string,
	start, end time.Time,
	guestEmail, guestFirstName, guestLastName, guestPhone string,
	adults, children int,
) (string, error) {
	return m.createResult, m.createErr
}
func (m *mockBookingService) GetReservationByCode(ctx context.Context, code string) (*domain.Reservation, error) {
	return m.getResult, m.getErr
}
func (m *mockBookingService) CancelReservation(ctx context.Context, id string) error {
	return m.cancelErr
}
func (m *mockBookingService) PreviewCancellation(ctx context.Context, id string) (float64, error) {
	return m.previewAmt, m.previewErr
}
