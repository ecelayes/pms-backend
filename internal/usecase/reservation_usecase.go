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
	db             *pgxpool.Pool
	roomRepo       *repository.RoomRepository
	resRepo        *repository.ReservationRepository
	guestRepo      *repository.GuestRepository
	ratePlanRepo   *repository.RatePlanRepository
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
		return "", fmt.Errorf("%w: at least 1 adult is required", entity.ErrInvalidInput)
	}
	if req.Children < 0 {
		return "", fmt.Errorf("%w: children cannot be negative", entity.ErrInvalidInput)
	}

	roomType, err := uc.roomRepo.GetByID(ctx, req.RoomTypeID)
	if err != nil {
		return "", entity.ErrRoomTypeNotFound
	}

	if req.Adults > roomType.MaxAdults {
		return "", fmt.Errorf("%w: exceeds max adults for this room", entity.ErrInvalidInput)
	}
	if req.Children > roomType.MaxChildren {
		return "", fmt.Errorf("%w: exceeds max children for this room", entity.ErrInvalidInput)
	}
	if (req.Adults + req.Children) > roomType.MaxOccupancy {
		return "", fmt.Errorf("%w: exceeds max total occupancy for this room", entity.ErrInvalidInput)
	}

	hotelCode, roomCode, err := uc.roomRepo.GetCodesForGeneration(ctx, req.RoomTypeID)
	if err != nil {
		return "", entity.ErrRoomTypeNotFound
	}
	resCode := fmt.Sprintf("%s-%s-%s", hotelCode, roomCode, utils.GenerateRandomCode(4))

	dailyRates, baseTotal, err := uc.pricingService.CalculateBaseRates(
		ctx,
		req.RoomTypeID,
		roomType.BasePrice,
		start,
		end,
	)
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
		
		if req.GuestFirstName != "" && req.GuestFirstName != guest.FirstName { needsUpdate = true }
		if req.GuestLastName != "" && req.GuestLastName != guest.LastName { needsUpdate = true }
		if req.GuestPhone != "" && req.GuestPhone != guest.Phone { needsUpdate = true }

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
		if err != nil { return "", fmt.Errorf("failed to generate uuid v7: %w", err) }
		
		newGuest := entity.Guest{
			BaseEntity: entity.BaseEntity{ ID: newID.String() },
			Email:     req.GuestEmail,
			FirstName: req.GuestFirstName,
			LastName:  req.GuestLastName,
			Phone:     req.GuestPhone,
		}
		
		guestID, err = uc.guestRepo.Create(ctx, tx, newGuest)
		if err != nil { return "", err }
	}
	
	lockedRoom, err := uc.roomRepo.GetByIDLocked(ctx, tx, req.RoomTypeID)
	if err != nil {
		return "", fmt.Errorf("failed to lock room inventory: %w", err)
	}

	reservedCount, err := uc.roomRepo.CountReservations(ctx, tx, req.RoomTypeID, start, end)
	if err != nil {
		return "", err
	}

	if (lockedRoom.TotalQuantity - reservedCount) < 1 {
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

func (uc *ReservationUseCase) PreviewCancellation(ctx context.Context, reservationID string) (float64, error) {
	res, err := uc.resRepo.GetByID(ctx, reservationID)
	if err != nil {
		return 0, err
	}

	if res.Status == "cancelled" {
		return 0, entity.ErrReservationCancelled
	}

	if res.RatePlanID == nil {
		return 0, nil
	}

	plan, err := uc.ratePlanRepo.GetByID(ctx, *res.RatePlanID)
	if err != nil {
		return 0, fmt.Errorf("failed to load rate plan policy: %w", err)
	}

	room, err := uc.roomRepo.GetByID(ctx, res.RoomTypeID)
	if err != nil {
		return 0, fmt.Errorf("failed to load room type: %w", err)
	}

	checkInTime := time.Date(res.Start.Year(), res.Start.Month(), res.Start.Day(), 15, 0, 0, 0, time.UTC)
	hoursUntil := time.Until(checkInTime).Hours()

	firstNightPrice := 0.0
	
	firstNightRates, baseFirstNight, err := uc.pricingService.CalculateBaseRates(
		ctx, 
		res.RoomTypeID, 
		room.BasePrice,
		res.Start, 
		res.Start.AddDate(0, 0, 1),
	)
	
	if err == nil && len(firstNightRates) > 0 {
		totalPax := res.Adults + res.Children
		firstNightPrice = uc.pricingService.ApplyRatePlan(baseFirstNight, *plan, totalPax, 1)
	} else {
		days := int(res.End.Sub(res.Start).Hours() / 24)
		if days > 0 {
			firstNightPrice = res.TotalPrice / float64(days)
		}
	}

	penalty := plan.CancellationPolicy.CalculatePenaltyAmount(res.TotalPrice, firstNightPrice, hoursUntil)

	return penalty, nil
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
