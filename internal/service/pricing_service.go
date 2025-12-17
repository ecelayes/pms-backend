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

func (s *PricingService) CalculateBaseRates(ctx context.Context, roomTypeID string, fallbackPrice float64, start, end time.Time) ([]entity.DailyRate, float64, error) {
	rules, _, err := s.priceRepo.ListByRoomType(ctx, roomTypeID, entity.PaginationRequest{Page: 1, Limit: 1000})
	if err != nil {
		return nil, 0, err
	}

	var dailyRates []entity.DailyRate
	var total float64

	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		currentPrice := 0.0
		priceFound := false

		for _, rule := range rules {
			if (d.Equal(rule.Start) || d.After(rule.Start)) && d.Before(rule.End) {
				currentPrice = rule.Price
				priceFound = true
				break
			}
		}

		if !priceFound {
			if fallbackPrice > 0 {
				currentPrice = fallbackPrice
				priceFound = true
			} else {
				return nil, 0, errors.New("no price defined for date: " + d.Format("2006-01-02"))
			}
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

	// Future logic: Discounts, Taxes, etc. can be added here centrally.

	return finalTotal
}
