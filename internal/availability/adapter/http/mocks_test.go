package http

import (
	"context"
	"time"

	"github.com/ecelayes/pms-backend/internal/shared/dto"
)

type mockAvailabilityService struct {
	searchResult        []dto.AvailabilityResult
	searchErr           error
	updateInventoryErr  error
}

func (m *mockAvailabilityService) Search(ctx context.Context, propertyID string, start, end time.Time, adults, children, rooms int) ([]dto.AvailabilityResult, error) {
	return m.searchResult, m.searchErr
}
func (m *mockAvailabilityService) UpdateInventory(ctx context.Context, propertyID, unitTypeID string, start, end time.Time, delta int) error {
	return m.updateInventoryErr
}
