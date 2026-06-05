package application

import (
	"context"
	"errors"
	"github.com/ecelayes/pms-backend/internal/pricing/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"time"
)

var (
	ErrNotFound = domain.ErrNotFound
)

type PricingService struct {
	priceRepo    domain.PriceRuleRepository
	ratePlanRepo domain.RatePlanRepository
}

func NewPricingService(pr domain.PriceRuleRepository, rr domain.RatePlanRepository) *PricingService {
	return &PricingService{
		priceRepo:    pr,
		ratePlanRepo: rr,
	}
}
func (s *PricingService) CreateRatePlan(
	ctx context.Context,
	propertyID, name string,
	meal domain.MealPlan,
	cancel domain.CancellationPolicy,
	pay domain.PaymentPolicy,
) (string, error) {
	rp, err := domain.NewRatePlan(propertyID, name, meal, cancel, pay)
	if err != nil {
		return "", err
	}
	if err := s.ratePlanRepo.Save(ctx, rp); err != nil {
		return "", err
	}
	return rp.ID(), nil
}
func (s *PricingService) GetRatePlan(ctx context.Context, id string) (*domain.RatePlan, error) {
	rp, err := s.ratePlanRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rp == nil {
		return nil, ErrNotFound
	}
	return rp, nil
}
func (s *PricingService) ListRatePlans(ctx context.Context, propertyID string) ([]*domain.RatePlan, error) {
	return s.ratePlanRepo.FindByPropertyID(ctx, propertyID)
}
func (s *PricingService) UpdateRatePlan(ctx context.Context, id string, name, desc string, active bool) error {
	rp, err := s.GetRatePlan(ctx, id)
	if err != nil {
		return err
	}
	rp.Update(name, desc, active)
	return s.ratePlanRepo.Save(ctx, rp)
}
func (s *PricingService) DeleteRatePlan(ctx context.Context, id string) error {
	return s.ratePlanRepo.Delete(ctx, id)
}
func (s *PricingService) CalculateBasePrice(ctx context.Context, unitTypeID string, defaultPrice vo.Money, start, end time.Time) (vo.Money, error) {
	total := vo.NewMoney(0, defaultPrice.Currency())
	startStr := start.Format("2006-01-02")
	endStr := end.Format("2006-01-02")
	rules, err := s.priceRepo.FindOverlapping(ctx, unitTypeID, startStr, endStr)
	if err != nil {
		return vo.Money{}, err
	}
	priceMap := make(map[string]vo.Money)
	for _, rule := range rules {
		curr := rule.DateRange().Start()
		ruleEnd := rule.DateRange().End()
		for curr.Before(ruleEnd) {
			priceMap[curr.Format("2006-01-02")] = rule.Price()
			curr = curr.AddDate(0, 0, 1)
		}
	}
	curr := start
	for curr.Before(end) {
		dayStr := curr.Format("2006-01-02")
		dayPrice, ok := priceMap[dayStr]
		if !ok {
			dayPrice = defaultPrice
		}
		total, err = total.Add(dayPrice)
		if err != nil {
			return vo.Money{}, err
		}
		curr = curr.AddDate(0, 0, 1)
	}
	return total, nil
}
func (s *PricingService) CalculateStayPrice(
	ctx context.Context,
	unitTypeID string,
	ratePlanID string,
	defaultPrice vo.Money,
	start, end time.Time,
	adults, children int,
) (vo.Money, error) {
	base, err := s.CalculateBasePrice(ctx, unitTypeID, defaultPrice, start, end)
	if err != nil {
		return vo.Money{}, err
	}
	if ratePlanID != "" {
		rp, err := s.ratePlanRepo.FindByID(ctx, ratePlanID)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return vo.Money{}, err
		}
		if rp != nil {
			return rp.CalculateTotal(base, start, end, adults, children), nil
		}
	}
	return base, nil
}
func (s *PricingService) CalculateCancellationPenalty(
	ctx context.Context,
	ratePlanID string,
	reservationStart time.Time,
	cancelDate time.Time,
	totalPrice vo.Money,
) (vo.Money, error) {
	rp, err := s.ratePlanRepo.FindByID(ctx, ratePlanID)
	if err != nil {
		return vo.Money{}, err
	}
	if rp == nil {
		return vo.Money{}, ErrNotFound
	}
	policy := rp.CancellationPolicy()
	if !policy.IsRefundable {
		return totalPrice, nil
	}
	hoursBefore := int(reservationStart.Sub(cancelDate).Hours())
	if hoursBefore < 0 {
		return totalPrice, nil
	}
	matched := false
	var activeRule domain.CancellationRule
	for _, rule := range policy.Rules {
		if hoursBefore < rule.HoursBeforeCheckIn {
			if !matched {
				matched = true
				activeRule = rule
			} else if rule.HoursBeforeCheckIn < activeRule.HoursBeforeCheckIn {
				activeRule = rule
			}
		}
	}
	if !matched {
		return vo.NewMoney(0, totalPrice.Currency()), nil
	}
	var amount int64
	switch activeRule.PenaltyType {
	case domain.PenaltyFixedAmount:
		amount = activeRule.PenaltyValue
	case domain.PenaltyPercentage:
		amount = totalPrice.Amount() * activeRule.PenaltyValue / 100
	case domain.PenaltyNights:
		return vo.Money{}, errors.New("nights penalty not supported yet")
	}
	return vo.NewMoney(amount, totalPrice.Currency()), nil
}
func (s *PricingService) SetPriceRule(ctx context.Context, unitTypeID string, start, end time.Time, price vo.Money) error {
	newRange, err := vo.NewDateRange(start, end)
	if err != nil {
		return err
	}
	overlaps, err := s.priceRepo.FindOverlapping(ctx, unitTypeID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return err
	}
	for _, existing := range overlaps {
		if err := s.priceRepo.Delete(ctx, existing.ID()); err != nil {
			return err
		}
		if existing.DateRange().Start().Before(newRange.Start()) {
			leftEnd := newRange.Start().AddDate(0, 0, -1)
			leftRange, err := vo.NewDateRange(existing.DateRange().Start(), leftEnd)
			if err == nil {
				leftRule := domain.NewPriceRule(unitTypeID, leftRange, existing.Price())
				if err := s.priceRepo.Save(ctx, leftRule); err != nil {
					return err
				}
			}
		}
		if existing.DateRange().End().After(newRange.End()) {
			rightStart := newRange.End().AddDate(0, 0, 1)
			rightRange, err := vo.NewDateRange(rightStart, existing.DateRange().End())
			if err == nil {
				rightRule := domain.NewPriceRule(unitTypeID, rightRange, existing.Price())
				if err := s.priceRepo.Save(ctx, rightRule); err != nil {
					return err
				}
			}
		}
	}
	rule := domain.NewPriceRule(unitTypeID, newRange, price)
	return s.priceRepo.Save(ctx, rule)
}
func (s *PricingService) GetRules(ctx context.Context, propertyID, unitTypeID string) ([]*domain.PriceRule, error) {
	if unitTypeID != "" {
		return s.priceRepo.FindByUnitType(ctx, unitTypeID)
	}
	if propertyID != "" {
		return s.priceRepo.FindByPropertyID(ctx, propertyID)
	}
	return nil, nil
}
func (s *PricingService) DeleteRule(ctx context.Context, id string) error {
	return s.priceRepo.Delete(ctx, id)
}
