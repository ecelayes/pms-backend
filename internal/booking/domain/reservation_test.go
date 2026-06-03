package domain

import (
	"testing"
	"time"

	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"github.com/stretchr/testify/assert"
)

func TestReservation_NewReservation(t *testing.T) {
	futureStart := time.Now().AddDate(0, 0, 7)
	futureEnd := time.Now().AddDate(0, 0, 14)
	dateRange, _ := vo.NewDateRange(futureStart, futureEnd)
	price := vo.NewMoney(50000, "USD")

	t.Run("valid reservation", func(t *testing.T) {
		res, err := NewReservation(
			"prop-123",
			"unit-type-456",
			"rate-plan-789",
			"guest-001",
			dateRange,
			price,
			"guest@example.com",
		)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "prop-123", res.PropertyID())
		assert.Equal(t, "unit-type-456", res.UnitTypeID())
		assert.Equal(t, StatusPending, res.Status())
		assert.NotEmpty(t, res.ID())
		assert.NotEmpty(t, res.ReservationCode())
	})

	t.Run("rejects past dates", func(t *testing.T) {
		pastStart := time.Now().AddDate(0, 0, -1)
		pastEnd := time.Now().AddDate(0, 0, 5)
		pastRange, _ := vo.NewDateRange(pastStart, pastEnd)

		res, err := NewReservation(
			"prop-123",
			"unit-type-456",
			"rate-plan-789",
			"guest-001",
			pastRange,
			price,
			"guest@example.com",
		)

		assert.ErrorIs(t, err, ErrInvalidReservationDates)
		assert.Nil(t, res)
	})
}

func TestReservation_Reconstitute(t *testing.T) {
	futureStart := time.Now().AddDate(0, 0, 7)
	futureEnd := time.Now().AddDate(0, 0, 14)
	dateRange, _ := vo.NewDateRange(futureStart, futureEnd)
	price := vo.NewMoney(50000, "USD")

	res := Reconstitute(
		"res-123",
		"prop-123",
		"unit-type-456",
		"rate-plan-789",
		"unit-001",
		"guest-001",
		dateRange,
		price,
		StatusConfirmed,
		"guest@example.com",
		"ABC12345",
		time.Now(),
	)

	assert.Equal(t, "res-123", res.ID())
	assert.Equal(t, "prop-123", res.PropertyID())
	assert.Equal(t, "unit-001", res.UnitID())
	assert.Equal(t, StatusConfirmed, res.Status())
	assert.Equal(t, "ABC12345", res.ReservationCode())
	
	// Reconstitute should not publish events
	events := res.PopEvents()
	assert.Empty(t, events)
}

func TestReservation_Confirm(t *testing.T) {
	futureStart := time.Now().AddDate(0, 0, 7)
	futureEnd := time.Now().AddDate(0, 0, 14)
	dateRange, _ := vo.NewDateRange(futureStart, futureEnd)
	price := vo.NewMoney(50000, "USD")

	t.Run("confirm pending reservation", func(t *testing.T) {
		res, _ := NewReservation(
			"prop-123", "unit-type-456", "rate-plan-789", "guest-001",
			dateRange, price, "guest@example.com",
		)

		assert.Equal(t, StatusPending, res.Status())
		
		err := res.Confirm()
		assert.NoError(t, err)
		assert.Equal(t, StatusConfirmed, res.Status())
	})

	t.Run("reject double confirm", func(t *testing.T) {
		res, _ := NewReservation(
			"prop-123", "unit-type-456", "rate-plan-789", "guest-001",
			dateRange, price, "guest@example.com",
		)

		res.Confirm()
		err := res.Confirm()
		
		assert.ErrorIs(t, err, ErrReservationAlreadyConfirmed)
	})
}

func TestReservation_Cancel(t *testing.T) {
	futureStart := time.Now().AddDate(0, 0, 7)
	futureEnd := time.Now().AddDate(0, 0, 14)
	dateRange, _ := vo.NewDateRange(futureStart, futureEnd)
	price := vo.NewMoney(50000, "USD")

	t.Run("cancel pending reservation", func(t *testing.T) {
		res, _ := NewReservation(
			"prop-123", "unit-type-456", "rate-plan-789", "guest-001",
			dateRange, price, "guest@example.com",
		)

		res.Cancel()
		assert.Equal(t, StatusCancelled, res.Status())
	})

	t.Run("cancel confirmed reservation", func(t *testing.T) {
		res, _ := NewReservation(
			"prop-123", "unit-type-456", "rate-plan-789", "guest-001",
			dateRange, price, "guest@example.com",
		)
		res.Confirm()

		res.Cancel()
		assert.Equal(t, StatusCancelled, res.Status())
	})
}

func TestReservation_Events(t *testing.T) {
	futureStart := time.Now().AddDate(0, 0, 7)
	futureEnd := time.Now().AddDate(0, 0, 14)
	dateRange, _ := vo.NewDateRange(futureStart, futureEnd)
	price := vo.NewMoney(50000, "USD")

	t.Run("new reservation publishes created event", func(t *testing.T) {
		res, _ := NewReservation(
			"prop-123", "unit-type-456", "rate-plan-789", "guest-001",
			dateRange, price, "guest@example.com",
		)

		events := res.PopEvents()
		assert.Len(t, events, 1)
		
		createdEvent, ok := events[0].(*ReservationCreatedEvent)
		assert.True(t, ok)
		assert.Equal(t, "reservation.created", createdEvent.Name())
		assert.Equal(t, res.ID(), createdEvent.ReservationID)
		assert.Equal(t, "prop-123", createdEvent.PropertyID)
	})

	t.Run("confirm publishes confirmed event", func(t *testing.T) {
		res, _ := NewReservation(
			"prop-123", "unit-type-456", "rate-plan-789", "guest-001",
			dateRange, price, "guest@example.com",
		)
		res.PopEvents() // Clear initial events

		res.Confirm()
		
		events := res.PopEvents()
		assert.Len(t, events, 1)
		
		confirmedEvent, ok := events[0].(*ReservationConfirmedEvent)
		assert.True(t, ok)
		assert.Equal(t, "reservation.confirmed", confirmedEvent.Name())
	})

	t.Run("cancel publishes cancelled event", func(t *testing.T) {
		res, _ := NewReservation(
			"prop-123", "unit-type-456", "rate-plan-789", "guest-001",
			dateRange, price, "guest@example.com",
		)
		res.PopEvents() // Clear initial events

		res.Cancel()
		
		events := res.PopEvents()
		assert.Len(t, events, 1)
		
		cancelledEvent, ok := events[0].(*ReservationCancelledEvent)
		assert.True(t, ok)
		assert.Equal(t, "reservation.cancelled", cancelledEvent.Name())
	})

	t.Run("popEvents clears the queue", func(t *testing.T) {
		res, _ := NewReservation(
			"prop-123", "unit-type-456", "rate-plan-789", "guest-001",
			dateRange, price, "guest@example.com",
		)
		
		res.PopEvents()
		events := res.PopEvents()
		assert.Empty(t, events)
	})
}
