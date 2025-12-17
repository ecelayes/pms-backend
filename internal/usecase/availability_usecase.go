package usecase

import (
	"context"
	"math"
	"time"

	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
	"github.com/ecelayes/pms-backend/internal/service"
)

type AvailabilityUseCase struct {
	roomRepo     *repository.RoomRepository
	resRepo      *repository.ReservationRepository
	ratePlanRepo *repository.RatePlanRepository
	pricingService *service.PricingService
}

func NewAvailabilityUseCase(
	roomRepo *repository.RoomRepository,
	resRepo *repository.ReservationRepository,
	ratePlanRepo *repository.RatePlanRepository,
	pricingService *service.PricingService,
) *AvailabilityUseCase {
	return &AvailabilityUseCase{
		roomRepo:       roomRepo,
		resRepo:        resRepo,
		ratePlanRepo:   ratePlanRepo,
		pricingService: pricingService,
	}
}

func (uc *AvailabilityUseCase) Search(ctx context.Context, filter entity.AvailabilityFilter) ([]entity.AvailabilitySearch, int64, error) {
	if !filter.End.After(filter.Start) {
		return nil, 0, entity.ErrInvalidDateRange
	}

	var roomTypes []entity.RoomType
	var err error
	if filter.HotelID != "" {
		roomTypes, _, err = uc.roomRepo.ListByHotel(ctx, filter.HotelID, entity.PaginationRequest{Page: 1, Limit: 1000})
	} else {
		roomTypes, err = uc.roomRepo.GetAll(ctx)
	}
	if err != nil {
		return nil, 0, err
	}

	var ratePlans []entity.RatePlan
	if filter.HotelID != "" {
		ratePlans, _, err = uc.ratePlanRepo.ListByHotel(ctx, filter.HotelID, entity.PaginationRequest{Page: 1, Limit: 1000})
	} else {
		ratePlans, err = uc.ratePlanRepo.GetAll(ctx)
	}
	if err != nil {
		return nil, 0, err
	}

	var results []entity.AvailabilitySearch
	nights := int(filter.End.Sub(filter.Start).Hours() / 24)

	for _, rt := range roomTypes {
		roomsNeeded := filter.Rooms
		if roomsNeeded <= 0 {
			roomsNeeded = 1
		}

		totalPax := filter.Adults + filter.Children
		reqTotalPerRoom := int(math.Ceil(float64(totalPax) / float64(roomsNeeded)))

		if rt.MaxOccupancy < reqTotalPerRoom {
			continue
		}

		reservedCount, err := uc.resRepo.CountOverlapping(ctx, rt.ID, filter.Start, filter.End)
		if err != nil {
			return nil, 0, err
		}

		available := rt.TotalQuantity - reservedCount
		if available < roomsNeeded {
			continue
		}

		baseDailyRates, baseTotal, err := uc.pricingService.CalculateBaseRates(
			ctx,
			rt.ID,
			rt.BasePrice,
			filter.Start,
			filter.End,
		)
		if err != nil {
			continue 
		}

		var rateOptions []entity.RateOption

		for _, rp := range ratePlans {
			if rp.HotelID != rt.HotelID {
				continue
			}

			if rp.RoomTypeID != nil && *rp.RoomTypeID != rt.ID { 
				continue 
			}
			if !rp.Active { 
				continue 
			}

			finalTotal := uc.pricingService.ApplyRatePlan(baseTotal, rp, totalPax, nights)
			
			finalTotal *= float64(roomsNeeded)

			finalDailyRates := make([]entity.DailyRate, len(baseDailyRates))
			copy(finalDailyRates, baseDailyRates)

			rateOptions = append(rateOptions, entity.RateOption{
				RatePlanID:         rp.ID,
				RatePlanName:       rp.Name,
				Description:        rp.Description,
				TotalPrice:         finalTotal,
				CancellationPolicy: rp.CancellationPolicy,
				MealPlan:           rp.MealPlan,
				PaymentPolicy:      rp.PaymentPolicy,
				NightlyRates:       finalDailyRates,
			})
		}

		if len(rateOptions) > 0 {
			results = append(results, entity.AvailabilitySearch{
				RoomTypeID:   rt.ID,
				RoomTypeName: rt.Name,
				AvailableQty: rt.TotalQuantity - reservedCount,
				MaxOccupancy: rt.MaxOccupancy,
				MaxAdults:    rt.MaxAdults,
				MaxChildren:  rt.MaxChildren,
				Amenities:    rt.Amenities,
				Rates:        rateOptions,
			})
		}
	}

	totalItems := int64(len(results))
	
	page := filter.Page
	if page < 1 { page = 1 }
	limit := filter.Limit
	if limit < 1 { limit = 10 }

	start := (page - 1) * limit
	end := start + limit

	if start > int(totalItems) {
		start = int(totalItems)
	}
	if end > int(totalItems) {
		end = int(totalItems)
	}

	paginatedResults := results[start:end]

	return paginatedResults, totalItems, nil
}

func calculateStayPrice(start, end time.Time, rules []entity.PriceRule) ([]entity.DailyRate, float64, bool) {
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
			return nil, 0, false
		}

		dailyRates = append(dailyRates, entity.DailyRate{
			Date:  d.Format("2006-01-02"),
			Price: currentPrice,
		})
		total += currentPrice
	}

	return dailyRates, total, true
}
