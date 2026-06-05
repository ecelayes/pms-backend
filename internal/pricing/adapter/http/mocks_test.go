package http

import (
	"context"
	"time"

	"github.com/ecelayes/pms-backend/internal/pricing/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
)

type mockPricingService struct {
	setRuleErr      error
	getRulesResult  []*domain.PriceRule
	getRulesErr     error
	deleteRuleErr   error
	createRPResult  string
	createRPErr     error
	getRPResult     *domain.RatePlan
	getRPErr        error
	listRPResult    []*domain.RatePlan
	listRPErr       error
	updateRPErr     error
	deleteRPErr     error
}

func (m *mockPricingService) SetPriceRule(ctx context.Context, unitTypeID string, start, end time.Time, price vo.Money) error {
	return m.setRuleErr
}
func (m *mockPricingService) GetRules(ctx context.Context, propertyID, unitTypeID string) ([]*domain.PriceRule, error) {
	return m.getRulesResult, m.getRulesErr
}
func (m *mockPricingService) DeleteRule(ctx context.Context, id string) error {
	return m.deleteRuleErr
}
func (m *mockPricingService) CreateRatePlan(ctx context.Context, propertyID, name string, mealPlan domain.MealPlan, cancellationPolicy domain.CancellationPolicy, paymentPolicy domain.PaymentPolicy) (string, error) {
	return m.createRPResult, m.createRPErr
}
func (m *mockPricingService) GetRatePlan(ctx context.Context, id string) (*domain.RatePlan, error) {
	return m.getRPResult, m.getRPErr
}
func (m *mockPricingService) ListRatePlans(ctx context.Context, propertyID string) ([]*domain.RatePlan, error) {
	return m.listRPResult, m.listRPErr
}
func (m *mockPricingService) UpdateRatePlan(ctx context.Context, id string, name, desc string, active bool) error {
	return m.updateRPErr
}
func (m *mockPricingService) DeleteRatePlan(ctx context.Context, id string) error {
	return m.deleteRPErr
}
