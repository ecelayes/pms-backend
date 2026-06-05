package http

import (
	"context"
	"time"

	"github.com/ecelayes/pms-backend/internal/pricing/application"
	"github.com/ecelayes/pms-backend/internal/pricing/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
)

// PricingService defines price rule + rate plan operations exposed to the HTTP layer.
type PricingService interface {
	SetPriceRule(ctx context.Context, unitTypeID string, start, end time.Time, price vo.Money) error
	GetRules(ctx context.Context, propertyID, unitTypeID string) ([]*domain.PriceRule, error)
	DeleteRule(ctx context.Context, id string) error
	CreateRatePlan(ctx context.Context, propertyID, name string, mealPlan domain.MealPlan, cancellationPolicy domain.CancellationPolicy, paymentPolicy domain.PaymentPolicy) (string, error)
	GetRatePlan(ctx context.Context, id string) (*domain.RatePlan, error)
	ListRatePlans(ctx context.Context, propertyID string) ([]*domain.RatePlan, error)
	UpdateRatePlan(ctx context.Context, id string, name, desc string, active bool) error
	DeleteRatePlan(ctx context.Context, id string) error
}

var _ PricingService = (*application.PricingService)(nil)
