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
	"github.com/ecelayes/pms-backend/internal/service"
	"github.com/ecelayes/pms-backend/internal/utils"
)

type ReservationUseCase struct {
	db       *pgxpool.Pool
	roomRepo *repository.RoomRepository
	resRepo  *repository.ReservationRepository
	guestRepo *repository.GuestRepository
	ratePlanRepo *repository.RatePlanRepository
	pricingService *service.PricingService
}

func NewReservationUseCase(
	db *pgxpool.Pool,
	roomRepo *repository.RoomRepository,
	resRepo *repository.ReservationRepository,
	guestRepo *repository.GuestRepository,
	ratePlanRepo *repository.RatePlanRepository,
	pricingService *service.PricingService,
) *ReservationUseCase {
	return &ReservationUseCase{
		db:             db,
		roomRepo:       roomRepo,
		resRepo:        resRepo,
		guestRepo:      guestRepo,
		ratePlanRepo:   ratePlanRepo,
		pricingService: pricingService,
	}
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

	nights := int(end.Sub(start).Hours() / 24)

	if req.Adults <= 0 {
		return "", errors.New("at least 1 adult is required")
	}
	if req.Children < 0 {
		return "", errors.New("children cannot be negative")
	}

	roomType, err := uc.roomRepo.GetByID(ctx, req.RoomTypeID)
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
	resCode := fmt.Sprintf("%s-%s-%s", hotelCode, roomCode, utils.GenerateRandomCode(4))

	dailyRates, baseTotal, err := uc.pricingService.CalculateBaseRates(ctx, req.RoomTypeID, start, end)
	if err != nil {
		return "", entity.ErrNoAvailability 
	}
	
	if len(dailyRates) != nights {
		return "", entity.ErrNoAvailability
	}

	finalPrice := baseTotal

	if req.RatePlanID != nil && *req.RatePlanID != "" {
		rp, err := uc.ratePlanRepo.GetByID(ctx, *req.RatePlanID)
		if err != nil {
			return "", fmt.Errorf("invalid rate plan: %w", err)
		}
		
		if rp.RoomTypeID != nil && *rp.RoomTypeID != req.RoomTypeID {
			return "", errors.New("rate plan not applicable to this room type")
		}
		if !rp.Active {
			return "", errors.New("rate plan is not active")
		}

		totalPax := req.Adults + req.Children
		finalPrice = uc.pricingService.ApplyRatePlan(baseTotal, *rp, totalPax, nights)
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

		newID, err := uuid.NewV7()
		if err != nil {
			return "", fmt.Errorf("failed to generate uuid v7: %w", err)
		}
		
		newGuest := entity.Guest{
			BaseEntity: entity.BaseEntity{
				ID: newID.String(),
			},
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

	newID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid v7: %w", err)
	}

	res := entity.Reservation{
		BaseEntity: entity.BaseEntity{
			ID: newID.String(),
		},
		ReservationCode: resCode,
		RoomTypeID:      req.RoomTypeID,
		GuestID:         guestID,
		Start:           start,
		End:             end,
		RatePlanID:      req.RatePlanID,
		TotalPrice:      finalPrice,
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
