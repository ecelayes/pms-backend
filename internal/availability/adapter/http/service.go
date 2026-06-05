package http

import (
	"context"
	"time"

	"github.com/ecelayes/pms-backend/internal/availability/application"
	"github.com/ecelayes/pms-backend/internal/shared/dto"
)

// AvailabilityService defines availability operations exposed to the HTTP layer.
type AvailabilityService interface {
	Search(ctx context.Context, propertyID string, start, end time.Time, adults, children, rooms int) ([]dto.AvailabilityResult, error)
	UpdateInventory(ctx context.Context, propertyID, unitTypeID string, start, end time.Time, delta int) error
}

var _ AvailabilityService = (*application.AvailabilityService)(nil)
