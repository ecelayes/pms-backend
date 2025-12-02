package usecase

import (
	"context"
	"time"

	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
)

type AvailabilityUseCase struct {
	repo *repository.RoomRepository
}

func NewAvailabilityUseCase(repo *repository.RoomRepository) *AvailabilityUseCase {
	return &AvailabilityUseCase{repo: repo}
}

func (uc *AvailabilityUseCase) Search(ctx context.Context, start, end time.Time) ([]entity.AvailabilitySearch, error) {
	roomTypes, err := uc.repo.GetAllRoomTypes(ctx)
	if err != nil {
		return nil, err
	}

	var results []entity.AvailabilitySearch

	for _, rt := range roomTypes {
		reservedCount, err := uc.repo.CountReservations(ctx, nil, rt.ID, start, end)
		if err != nil {
				return nil, err
		}

		available := rt.TotalQuantity - reservedCount
		if available <= 0 {
			continue
		}

		dailyRates, err := uc.repo.GetDailyPrices(ctx, rt.ID, start, end)
		if err != nil {
			return nil, err
		}

		daysRequested := int(end.Sub(start).Hours() / 24)
		if len(dailyRates) != daysRequested {
			// Missing prices for some dates -> Room is not bookable
			continue
		}

		var total float64
		for _, r := range dailyRates {
			total += r.Price
		}

		results = append(results, entity.AvailabilitySearch{
			RoomTypeID:   rt.ID,
			RoomTypeName: rt.Name,
			AvailableQty: available,
			TotalPrice:   total,
			NightlyRates: dailyRates,
		})
	}

	return results, nil
}
