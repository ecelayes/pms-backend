package usecase

import (
	"context"
	"math"
	"time"

	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
)

type AvailabilityUseCase struct {
	roomRepo  *repository.RoomRepository
	resRepo   *repository.ReservationRepository
	priceRepo *repository.PriceRepository
}

func NewAvailabilityUseCase(
	roomRepo *repository.RoomRepository,
	resRepo *repository.ReservationRepository,
	priceRepo *repository.PriceRepository,
) *AvailabilityUseCase {
	return &AvailabilityUseCase{
		roomRepo:  roomRepo,
		resRepo:   resRepo,
		priceRepo: priceRepo,
	}
}

func (uc *AvailabilityUseCase) Search(ctx context.Context, filter entity.AvailabilityFilter) ([]entity.AvailabilitySearch, error) {
	if !filter.End.After(filter.Start) {
		return nil, entity.ErrInvalidDateRange
	}

	var roomTypes []entity.RoomType
	var err error
	if filter.HotelID != "" {
		roomTypes, err = uc.roomRepo.ListByHotel(ctx, filter.HotelID)
	} else {
		roomTypes, err = uc.roomRepo.GetAll(ctx)
	}
	if err != nil { return nil, err }

	var results []entity.AvailabilitySearch

	for _, rt := range roomTypes {
		roomsNeeded := filter.Rooms
		if roomsNeeded <= 0 { roomsNeeded = 1 }

		reqAdultsPerRoom := int(math.Ceil(float64(filter.Adults) / float64(roomsNeeded)))
		reqChildrenPerRoom := int(math.Ceil(float64(filter.Children) / float64(roomsNeeded)))
		reqTotalPerRoom := int(math.Ceil(float64(filter.Adults+filter.Children) / float64(roomsNeeded)))

		if rt.MaxAdults < reqAdultsPerRoom { continue }
		if rt.MaxChildren < reqChildrenPerRoom { continue }
		if rt.MaxOccupancy < reqTotalPerRoom { continue }

		reservedCount, err := uc.resRepo.CountOverlapping(ctx, rt.ID, filter.Start, filter.End)
		if err != nil { return nil, err }

		available := rt.TotalQuantity - reservedCount
		if available < roomsNeeded {
			continue
		}

		priceRules, err := uc.priceRepo.ListByRoomType(ctx, rt.ID)
		if err != nil { return nil, err }

		dailyRates, totalPrice, isCovered := calculateStayPrice(filter.Start, filter.End, priceRules)
		
		if !isCovered {
			continue
		}

		totalPrice *= float64(roomsNeeded)

		results = append(results, entity.AvailabilitySearch{
			RoomTypeID:   rt.ID,
			RoomTypeName: rt.Name,
			AvailableQty: available,
			TotalPrice:   totalPrice,
			MaxOccupancy: rt.MaxOccupancy,
			MaxAdults:    rt.MaxAdults,
			MaxChildren:  rt.MaxChildren,
			Amenities:    rt.Amenities,
			NightlyRates: dailyRates,
		})
	}

	return results, nil
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
