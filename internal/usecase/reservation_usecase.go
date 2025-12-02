package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
)

type ReservationUseCase struct {
	db       *pgxpool.Pool
	roomRepo *repository.RoomRepository
}

func NewReservationUseCase(db *pgxpool.Pool, roomRepo *repository.RoomRepository) *ReservationUseCase {
	return &ReservationUseCase{db: db, roomRepo: roomRepo}
}

func (uc *ReservationUseCase) Create(ctx context.Context, req entity.CreateReservationRequest) (string, error) {
	layout := "2006-01-02"
	start, err := time.Parse(layout, req.Start)
	if err != nil {
		return "", entity.ErrInvalidDateFormat
	}
	end, err := time.Parse(layout, req.End)
	if err != nil {
		return "", entity.ErrInvalidDateFormat
	}
	if !end.After(start) {
		return "", entity.ErrInvalidDateRange
	}

	dailyRates, err := uc.roomRepo.GetDailyPrices(ctx, req.RoomTypeID, start, end)
	if err != nil {
		return "", err
	}
	
	days := int(end.Sub(start).Hours() / 24)
	if len(dailyRates) != days {
		return "", errors.New("price not available for selected dates")
	}

	var totalPrice float64
	for _, rate := range dailyRates {
		totalPrice += rate.Price
	}

	tx, err := uc.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	reservedCount, err := uc.roomRepo.CountReservations(ctx, tx, req.RoomTypeID, start, end)
	if err != nil {
		return "", err
	}

	roomType, err := uc.roomRepo.GetRoomTypeByID(ctx, req.RoomTypeID) 
	if err != nil {
		return "", errors.New("room type not found")
	}

	if (roomType.TotalQuantity - reservedCount) <= 0 {
		return "", entity.ErrNoAvailability
	}

	resID := uuid.New().String()
	reservation := entity.Reservation{
		ID:         resID,
		RoomTypeID: req.RoomTypeID,
		GuestEmail: req.GuestEmail,
		Start:      start,
		End:        end,
		TotalPrice: totalPrice,
	}

	if err := uc.roomRepo.CreateReservation(ctx, tx, reservation); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}

	return resID, nil
}
