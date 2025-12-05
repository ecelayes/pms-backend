package usecase

import (
	"context"
	"math"

	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
)

type AvailabilityUseCase struct {
	repo *repository.RoomRepository
}

func NewAvailabilityUseCase(repo *repository.RoomRepository) *AvailabilityUseCase {
	return &AvailabilityUseCase{repo: repo}
}

func (uc *AvailabilityUseCase) Search(ctx context.Context, filter entity.AvailabilityFilter) ([]entity.AvailabilitySearch, error) {
	if !filter.End.After(filter.Start) {
		return nil, entity.ErrInvalidDateRange
	}

	roomTypes, err := uc.repo.GetAllRoomTypes(ctx, filter.HotelID)
	if err != nil { return nil, err }

	var results []entity.AvailabilitySearch

	for _, rt := range roomTypes {
		roomsNeeded := filter.Rooms
		if roomsNeeded <= 0 { roomsNeeded = 1 }

		reqAdultsPerRoom := int(math.Ceil(float64(filter.Adults) / float64(roomsNeeded)))
		reqChildrenPerRoom := int(math.Ceil(float64(filter.Children) / float64(roomsNeeded)))
		
		reqTotalPerRoom := int(math.Ceil(float64(filter.Adults + filter.Children) / float64(roomsNeeded)))

		if rt.MaxAdults < reqAdultsPerRoom { continue }
		if rt.MaxChildren < reqChildrenPerRoom { continue }
		if rt.MaxOccupancy < reqTotalPerRoom { continue }

		reservedCount, err := uc.repo.CountReservations(ctx, nil, rt.ID, filter.Start, filter.End)
		if err != nil { return nil, err }

		available := rt.TotalQuantity - reservedCount
		
		if available < roomsNeeded {
			continue
		}

		dailyRates, err := uc.repo.GetDailyPrices(ctx, rt.ID, filter.Start, filter.End)
		if err != nil { return nil, err }

		expectedNights := int(filter.End.Sub(filter.Start).Hours() / 24)
		if len(dailyRates) != expectedNights { continue }

		var totalPrice float64
		for _, r := range dailyRates {
			totalPrice += r.Price
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
