package application

import (
	"context"
	"time"

	"github.com/ecelayes/pms-backend/internal/catalog/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
)

// CatalogPort defines the interface for catalog operations needed by availability
type CatalogPort interface {
	ListUnitTypes(ctx context.Context, propertyID string, page, limit int) ([]*domain.UnitType, int64, error)
}

// PricingPort defines the interface for pricing operations needed by availability
type PricingPort interface {
	CalculateBasePrice(ctx context.Context, unitTypeID string, defaultPrice vo.Money, start, end time.Time) (vo.Money, error)
}
