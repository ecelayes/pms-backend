package domain

import (
	"context"
	"time"
)

type AvailabilityRepository interface {
	UpdateInventory(ctx context.Context, propertyID string, unitID string, date time.Time, delta int) error
	GetInventory(ctx context.Context, propertyID string, unitID string, date time.Time) (int, error)
	GetBatchInventory(ctx context.Context, propertyID string, unitID string, dates []time.Time) (map[string]int, error)
}
