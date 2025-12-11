package service

import (
	"context"
	"errors"
	"time"

	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
)

type PricingService struct {
	priceRepo *repository.PriceRepository
}

func NewPricingService(priceRepo *repository.PriceRepository) *PricingService {
	return &PricingService{priceRepo: priceRepo}
}

func (s *PricingService) CalculateBaseRates(ctx context.Context, roomTypeID string, start, end time.Time) ([]entity.DailyRate, float64, error) {
	rules, err := s.priceRepo.ListByRoomType(ctx, roomTypeID)
	if err != nil {
		return nil, 0, err
	}

	var dailyRates []entity.DailyRate
	var total float64

	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		priceFound := false
		var currentPrice float64

		for _, rule := range rules {
			if (d.Equal(rule.Start) || d.After(rule.Start)) && d.Before(rule.End) {
				currentPrice = rule.Price
				priceFound = true
				break
			}
		}

		if !priceFound {
			return nil, 0, errors.New("no price defined for date: " + d.Format("2006-01-02"))
		}

		dailyRates = append(dailyRates, entity.DailyRate{
			Date:  d.Format("2006-01-02"),
			Price: currentPrice,
		})
		total += currentPrice
	}

	return dailyRates, total, nil
}

func (s *PricingService) ApplyRatePlan(baseTotal float64, plan entity.RatePlan, pax int, nights int) float64 {
	finalTotal := baseTotal

	if plan.MealPlan.Included && plan.MealPlan.PricePerPax > 0 {
		mealCost := plan.MealPlan.PricePerPax * float64(pax) * float64(nights)
		finalTotal += mealCost
	}

	// 2. Future logic: Discounts, Taxes, etc. can be added here centrally.

	return finalTotal
}
