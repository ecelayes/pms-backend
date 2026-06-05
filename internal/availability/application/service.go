package application

import (
	"context"
	"github.com/ecelayes/pms-backend/internal/availability/domain"
	"github.com/ecelayes/pms-backend/internal/shared/dto"
	"time"
)

type AvailabilityService struct {
	repo    domain.AvailabilityRepository
	catalog CatalogPort
	pricing PricingPort
}

func NewAvailabilityService(
	repo domain.AvailabilityRepository,
	catalog CatalogPort,
	pricing PricingPort,
) *AvailabilityService {
	return &AvailabilityService{
		repo:    repo,
		catalog: catalog,
		pricing: pricing,
	}
}

func (s *AvailabilityService) Search(ctx context.Context, propertyID string, start, end time.Time, adults, children, rooms int) ([]dto.AvailabilityResult, error) {
	unitTypes, _, err := s.catalog.ListUnitTypes(ctx, propertyID, 1, 1000)
	if err != nil {
		return nil, err
	}
	var results []dto.AvailabilityResult
	var dates []time.Time
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d)
	}
	if len(dates) == 0 {
		return results, nil
	}
	for _, ut := range unitTypes {
		if ut.MaxAdults() < adults {
			continue
		}
		if (ut.MaxAdults() + ut.MaxChildren()) < (adults + children) {
			continue
		}
		counts, err := s.repo.GetBatchInventory(ctx, propertyID, ut.ID(), dates)
		if err != nil {
			continue
		}
		minQty := ut.TotalQuantity()
		available := true
		for _, d := range dates {
			dStr := d.Format("2006-01-02")
			qty, ok := counts[dStr]
			if !ok {
				qty = ut.TotalQuantity()
			}
			if qty < minQty {
				minQty = qty
			}
		}
		if minQty < rooms {
			available = false
		}
		if !available {
			continue
		}
		price, err := s.pricing.CalculateBasePrice(ctx, ut.ID(), ut.BasePrice(), start, end)
		if err != nil {
			continue
		}
		if price.Amount() <= 0 {
			continue
		}
		results = append(results, dto.AvailabilityResult{
			UnitTypeID:   ut.ID(),
			Name:         ut.Name(),
			AvailableQty: minQty,
			TotalPrice:   float64(price.Amount()) / 100.0,
			Currency:     price.Currency(),
			BasePrice:    float64(ut.BasePrice().Amount()) / 100.0,
		})
	}
	return results, nil
}

func (s *AvailabilityService) UpdateInventory(ctx context.Context, propertyID, unitTypeID string, start, end time.Time, delta int) error {
	var dates []time.Time
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d)
	}
	for _, date := range dates {
		if err := s.repo.UpdateInventory(ctx, propertyID, unitTypeID, date, delta); err != nil {
			return err
		}
	}
	return nil
}
