package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/repository"
)

type ReservationUseCase struct {
	db       *pgxpool.Pool
	roomRepo *repository.RoomRepository
	resRepo  *repository.ReservationRepository
}

func NewReservationUseCase(db *pgxpool.Pool, roomRepo *repository.RoomRepository, resRepo *repository.ReservationRepository) *ReservationUseCase {
	return &ReservationUseCase{
		db:       db,
		roomRepo: roomRepo,
		resRepo:  resRepo,
	}
}

func generateSuffix(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
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

	hotelCode, roomCode, err := uc.roomRepo.GetCodesForGeneration(ctx, req.RoomTypeID)
	if err != nil {
		return "", entity.ErrRoomTypeNotFound
	}
	resCode := fmt.Sprintf("%s-%s-%s", hotelCode, roomCode, generateSuffix(4))

	dailyRates, err := uc.roomRepo.GetDailyPrices(ctx, req.RoomTypeID, start, end)
	if err != nil {
		return "", err
	}

	if len(dailyRates) != int(end.Sub(start).Hours()/24) {
		return "", entity.ErrNoAvailability
	}

	var totalPrice float64
	for _, rate := range dailyRates {
		totalPrice += rate.Price
	}

	tx, err := uc.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	reservedCount, err := uc.roomRepo.CountReservations(ctx, tx, req.RoomTypeID, start, end)
	if err != nil {
		return "", err
	}

	roomType, err := uc.roomRepo.GetRoomTypeByID(ctx, req.RoomTypeID)
	if err != nil {
		return "", entity.ErrRoomTypeNotFound
	}

	if (roomType.TotalQuantity - reservedCount) <= 0 {
		return "", entity.ErrNoAvailability
	}

	newID := uuid.New().String()
	
	res := entity.Reservation{
		BaseEntity: entity.BaseEntity{
			ID: newID,
		},
		ReservationCode: resCode,
		RoomTypeID:      req.RoomTypeID,
		GuestEmail:      req.GuestEmail,
		Start:           start,
		End:             end,
		TotalPrice:      totalPrice,
		Status:          "confirmed",
	}

	if err := uc.resRepo.Create(ctx, tx, res); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return resCode, nil
}

func (uc *ReservationUseCase) Cancel(ctx context.Context, id string) error {
	return uc.resRepo.UpdateStatus(ctx, id, "cancelled")
}

func (uc *ReservationUseCase) Delete(ctx context.Context, id string) error {
	return uc.resRepo.Delete(ctx, id)
}

func (uc *ReservationUseCase) GetByCode(ctx context.Context, code string) (*entity.Reservation, error) {
	return uc.resRepo.GetByCode(ctx, code)
}
