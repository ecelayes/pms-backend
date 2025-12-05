package usecase

import (
	"context"
	"errors"
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
	guestRepo *repository.GuestRepository
}

func NewReservationUseCase(db *pgxpool.Pool, roomRepo *repository.RoomRepository, resRepo *repository.ReservationRepository, guestRepo *repository.GuestRepository) *ReservationUseCase {
	return &ReservationUseCase{
		db:        db,
		roomRepo:  roomRepo,
		resRepo:   resRepo,
		guestRepo: guestRepo,
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

	if req.Adults <= 0 {
		return "", errors.New("at least 1 adult is required")
	}
	if req.Children < 0 {
		return "", errors.New("children cannot be negative")
	}

	roomType, err := uc.roomRepo.GetRoomTypeByID(ctx, req.RoomTypeID)
	if err != nil {
		return "", entity.ErrRoomTypeNotFound
	}

	if req.Adults > roomType.MaxAdults {
		return "", errors.New("exceeds max adults for this room")
	}
	if req.Children > roomType.MaxChildren {
		return "", errors.New("exceeds max children for this room")
	}
	if (req.Adults + req.Children) > roomType.MaxOccupancy {
		return "", errors.New("exceeds max total occupancy for this room")
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

	expectedNights := int(end.Sub(start).Hours() / 24)
	if len(dailyRates) != expectedNights {
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

	if req.GuestEmail == "" {
		return "", errors.New("guest email is required")
	}

	guest, err := uc.guestRepo.GetByEmail(ctx, req.GuestEmail)
	if err != nil {
		return "", err
	}

	var guestID string

	if guest != nil {
		guestID = guest.ID

		needsUpdate := false
		
		if req.GuestFirstName != "" && req.GuestFirstName != guest.FirstName {
			needsUpdate = true
		}
		if req.GuestLastName != "" && req.GuestLastName != guest.LastName {
			needsUpdate = true
		}
		if req.GuestPhone != "" && req.GuestPhone != guest.Phone {
			needsUpdate = true
		}

		if needsUpdate {
			updatedGuest := entity.Guest{
				Email:     guest.Email,
				FirstName: req.GuestFirstName,
				LastName:  req.GuestLastName,
				Phone:     req.GuestPhone,
			}
			if err := uc.guestRepo.Update(ctx, tx, updatedGuest); err != nil {
				return "", err
			}
		}

	} else {
		if req.GuestFirstName == "" || req.GuestLastName == "" {
			return "", errors.New("guest name is required for new registration")
		}
		
		newGuest := entity.Guest{
			Email:     req.GuestEmail,
			FirstName: req.GuestFirstName,
			LastName:  req.GuestLastName,
			Phone:     req.GuestPhone,
		}
		
		guestID, err = uc.guestRepo.Create(ctx, tx, newGuest)
		if err != nil {
			return "", err
		}
	}

	reservedCount, err := uc.roomRepo.CountReservations(ctx, tx, req.RoomTypeID, start, end)
	if err != nil {
		return "", err
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
		GuestID:         guestID,
		Start:           start,
		End:             end,
		TotalPrice:      totalPrice,
		Status:          "confirmed",
		
		Adults:   req.Adults,
		Children: req.Children,
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
