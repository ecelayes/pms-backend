package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ecelayes/pms-backend/internal/booking/domain"
	sharedDomain "github.com/ecelayes/pms-backend/internal/shared/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
)


type mockReservationRepo struct {
	saveErr             error
	findByIDResult      *domain.Reservation
	findByIDErr         error
	updateErr           error
	countOverlappingCnt int
	countOverlappingErr error
	lockUnitTypeErr     error
	getByCodeResult     *domain.Reservation
	getByCodeErr        error
	runInTxFunc         func(ctx context.Context, fn func(context.Context) error) error
}

func (m *mockReservationRepo) Save(ctx context.Context, r *domain.Reservation) error {
	return m.saveErr
}
func (m *mockReservationRepo) FindByID(ctx context.Context, id string) (*domain.Reservation, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	return m.findByIDResult, nil
}
func (m *mockReservationRepo) Update(ctx context.Context, r *domain.Reservation) error {
	return m.updateErr
}
func (m *mockReservationRepo) CountOverlapping(ctx context.Context, unitTypeID string, start, end time.Time) (int, error) {
	if m.countOverlappingErr != nil {
		return 0, m.countOverlappingErr
	}
	return m.countOverlappingCnt, nil
}
func (m *mockReservationRepo) LockUnitType(ctx context.Context, unitTypeID string) error {
	return m.lockUnitTypeErr
}
func (m *mockReservationRepo) GetByCode(ctx context.Context, code string) (*domain.Reservation, error) {
	if m.getByCodeErr != nil {
		return nil, m.getByCodeErr
	}
	return m.getByCodeResult, nil
}
func (m *mockReservationRepo) RunInTransaction(ctx context.Context, fn func(context.Context) error) error {
	if m.runInTxFunc != nil {
		return m.runInTxFunc(ctx, fn)
	}
	return fn(ctx)
}

type mockPricingService struct {
	calculateBasePriceResult vo.Money
	calculateBasePriceErr     error
	calculateStayPriceResult  vo.Money
	calculateStayPriceErr     error
	calculatePenaltyResult    vo.Money
	calculatePenaltyErr       error
}

func (m *mockPricingService) CalculateBasePrice(ctx context.Context, unitTypeID string, defaultPrice vo.Money, start, end time.Time) (vo.Money, error) {
	if m.calculateBasePriceErr != nil {
		return vo.Money{}, m.calculateBasePriceErr
	}
	return m.calculateBasePriceResult, nil
}
func (m *mockPricingService) CalculateStayPrice(ctx context.Context, unitTypeID, ratePlanID string, defaultPrice vo.Money, start, end time.Time, adults, children int) (vo.Money, error) {
	if m.calculateStayPriceErr != nil {
		return vo.Money{}, m.calculateStayPriceErr
	}
	return m.calculateStayPriceResult, nil
}
func (m *mockPricingService) CalculateCancellationPenalty(ctx context.Context, ratePlanID string, reservationStart time.Time, cancelDate time.Time, totalPrice vo.Money) (vo.Money, error) {
	if m.calculatePenaltyErr != nil {
		return vo.Money{}, m.calculatePenaltyErr
	}
	return m.calculatePenaltyResult, nil
}

type mockCatalogService struct {
	getUnitTypePropertyID string
	getUnitTypeBasePrice  vo.Money
	getUnitTypeTotalQty   int
	getUnitTypeErr        error
}

func (m *mockCatalogService) GetUnitType(ctx context.Context, id string) (string, vo.Money, int, error) {
	if m.getUnitTypeErr != nil {
		return "", vo.Money{}, 0, m.getUnitTypeErr
	}
	return m.getUnitTypePropertyID, m.getUnitTypeBasePrice, m.getUnitTypeTotalQty, nil
}

type mockIdentityService struct {
	findOrCreateGuestID  string
	findOrCreateGuestErr error
}

func (m *mockIdentityService) FindOrCreateGuest(ctx context.Context, email, firstName, lastName, phone string) (string, error) {
	if m.findOrCreateGuestErr != nil {
		return "", m.findOrCreateGuestErr
	}
	return m.findOrCreateGuestID, nil
}

type mockAvailabilityService struct {
	updateInventoryErr error
}

func (m *mockAvailabilityService) UpdateInventory(ctx context.Context, propertyID, unitTypeID string, start, end time.Time, delta int) error {
	return m.updateInventoryErr
}

type mockStreamPublisher struct {
	publishCreatedErr   error
	publishCancelledErr error
}

func (m *mockStreamPublisher) PublishReservationCreated(ctx context.Context, payload sharedDomain.ReservationCreatedPayload) error {
	return m.publishCreatedErr
}
func (m *mockStreamPublisher) PublishReservationCancelled(ctx context.Context, payload sharedDomain.ReservationCancelledPayload) error {
	return m.publishCancelledErr
}

func futureDateRange() vo.DateRange {
	now := time.Now()
	start := now.AddDate(0, 1, 0)
	end := start.AddDate(0, 0, 2) // 2 nights
	dr, _ := vo.NewDateRange(start, end)
	return dr
}


func TestNewBookingService(t *testing.T) {
	repo := &mockReservationRepo{}
	svc := NewBookingService(repo, nil, nil, nil, nil)
	if svc == nil {
		t.Error("Expected non-nil service")
	}
	if svc.repo != repo {
		t.Error("Repo not set correctly")
	}
}

func TestNewBookingServiceWithPublisher(t *testing.T) {
	repo := &mockReservationRepo{}
	publisher := &mockStreamPublisher{}
	svc := NewBookingService(repo, nil, nil, nil, nil, publisher)
	if svc.publisher != publisher {
		t.Error("Publisher not set correctly")
	}
}

func TestNewBookingServiceNilPublisher(t *testing.T) {
	repo := &mockReservationRepo{}
	svc := NewBookingService(repo, nil, nil, nil, nil)
	if svc.publisher != nil {
		t.Error("Publisher should be nil when not provided")
	}
}

func TestBookingService_CreateReservation_InvalidDateRange(t *testing.T) {
	repo := &mockReservationRepo{}
	svc := NewBookingService(repo, nil, nil, nil, nil)
	
	start := time.Date(2024, 6, 5, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC) // Before start
	
	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err == nil {
		t.Error("Expected error for invalid date range")
	}
}

func TestBookingService_CreateReservation_IdentityServiceError(t *testing.T) {
	repo := &mockReservationRepo{}
	identity := &mockIdentityService{findOrCreateGuestErr: errors.New("identity error")}
	svc := NewBookingService(repo, nil, nil, identity, nil)
	
	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()
	
	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err == nil {
		t.Error("Expected error from identity service")
	}
}

func TestBookingService_CreateReservation_TransactionFailsOnLock(t *testing.T) {
	repo := &mockReservationRepo{
		lockUnitTypeErr: errors.New("lock failed"),
		runInTxFunc: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx) // Execute the function
		},
	}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-123"}
	svc := NewBookingService(repo, nil, nil, identity, nil)
	
	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()
	
	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err == nil {
		t.Error("Expected error from lock failure")
	}
}

func TestBookingService_CreateReservation_CatalogError(t *testing.T) {
	repo := &mockReservationRepo{
		runInTxFunc: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-123"}
	catalog := &mockCatalogService{getUnitTypeErr: errors.New("catalog error")}
	svc := NewBookingService(repo, nil, catalog, identity, nil)
	
	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()
	
	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err == nil {
		t.Error("Expected error from catalog service")
	}
}

func TestBookingService_CreateReservation_NoAvailability(t *testing.T) {
	repo := &mockReservationRepo{
		countOverlappingCnt: 5, // 5 already booked
		runInTxFunc: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-123"}
	catalog := &mockCatalogService{
		getUnitTypePropertyID: "prop-1",
		getUnitTypeBasePrice:  vo.NewMoney(10000, "USD"),
		getUnitTypeTotalQty:   5, // Only 5 total
	}
	svc := NewBookingService(repo, nil, catalog, identity, nil)
	
	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()
	
	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err != ErrNoAvailability {
		t.Errorf("Expected ErrNoAvailability, got: %v", err)
	}
}

func TestBookingService_CreateReservation_PricingError(t *testing.T) {
	repo := &mockReservationRepo{
		countOverlappingCnt: 1,
		runInTxFunc: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-123"}
	catalog := &mockCatalogService{
		getUnitTypePropertyID: "prop-1",
		getUnitTypeBasePrice:  vo.NewMoney(10000, "USD"),
		getUnitTypeTotalQty:   5,
	}
	pricing := &mockPricingService{calculateStayPriceErr: errors.New("pricing error")}
	svc := NewBookingService(repo, pricing, catalog, identity, nil)
	
	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()
	
	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err == nil {
		t.Error("Expected error from pricing service")
	}
}

func TestBookingService_CreateReservation_Success(t *testing.T) {
	repo := &mockReservationRepo{
		countOverlappingCnt: 1,
		runInTxFunc: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
		saveErr: nil,
		// Must provide GetByCode result for the async goroutine
		getByCodeResult: func() *domain.Reservation {
			dr, _ := vo.NewDateRange(time.Now().AddDate(0, 1, 0), time.Now().AddDate(0, 1, 2))
			return domain.Reconstitute(
				"res-new", "prop-1", "ut-1", "rp-1", "", "guest-123",
				dr, vo.NewMoney(20000, "USD"), domain.StatusConfirmed,
				"test@test.com", "RES123", time.Now(),
			)
		}(),
	}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-123"}
	catalog := &mockCatalogService{
		getUnitTypePropertyID: "prop-1",
		getUnitTypeBasePrice:  vo.NewMoney(10000, "USD"),
		getUnitTypeTotalQty:   5,
	}
	pricing := &mockPricingService{calculateStayPriceResult: vo.NewMoney(20000, "USD")}
	svc := NewBookingService(repo, pricing, catalog, identity, nil)
	
	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()
	
	code, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 2, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if code == "" {
		t.Error("Expected non-empty reservation code")
	}
}

func TestBookingService_GetReservationByCode(t *testing.T) {
	expectedRes := &domain.Reservation{}
	repo := &mockReservationRepo{getByCodeResult: expectedRes}
	svc := NewBookingService(repo, nil, nil, nil, nil)
	
	res, err := svc.GetReservationByCode(context.Background(), "RES-123")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res != expectedRes {
		t.Error("Wrong reservation returned")
	}
}

func TestBookingService_GetReservationByCode_NotFound(t *testing.T) {
	repo := &mockReservationRepo{getByCodeResult: nil}
	svc := NewBookingService(repo, nil, nil, nil, nil)
	
	res, err := svc.GetReservationByCode(context.Background(), "NON-EXISTENT")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res != nil {
		t.Error("Expected nil reservation")
	}
}

func TestBookingService_CancelReservation_NotFound(t *testing.T) {
	repo := &mockReservationRepo{findByIDResult: nil}
	svc := NewBookingService(repo, nil, nil, nil, nil)
	
	err := svc.CancelReservation(context.Background(), "non-existent")
	if err == nil {
		t.Error("Expected error for not found reservation")
	}
}

func TestBookingService_CancelReservation_Success(t *testing.T) {
	dateRange, _ := vo.NewDateRange(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 3, 23, 59, 59, 0, time.UTC))
	res := domain.Reconstitute(
		"res-1", "prop-1", "ut-1", "rp-1", "", "guest-1",
		dateRange,
		vo.NewMoney(20000, "USD"),
		domain.StatusConfirmed,
		"test@test.com", "RES123", time.Now(),
	)
	repo := &mockReservationRepo{findByIDResult: res}
	svc := NewBookingService(repo, nil, nil, nil, nil)
	
	err := svc.CancelReservation(context.Background(), "res-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res.Status() != domain.StatusCancelled {
		t.Errorf("Expected status cancelled, got: %s", res.Status())
	}
}

func TestBookingService_CancelReservation_UpdateError(t *testing.T) {
	dateRange, _ := vo.NewDateRange(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 3, 23, 59, 59, 0, time.UTC))
	res := domain.Reconstitute(
		"res-1", "prop-1", "ut-1", "rp-1", "", "guest-1",
		dateRange,
		vo.NewMoney(20000, "USD"),
		domain.StatusConfirmed,
		"test@test.com", "RES123", time.Now(),
	)
	repo := &mockReservationRepo{findByIDResult: res, updateErr: errors.New("update error")}
	svc := NewBookingService(repo, nil, nil, nil, nil)
	
	err := svc.CancelReservation(context.Background(), "res-1")
	if err == nil {
		t.Error("Expected error from update")
	}
}

func TestBookingService_PreviewCancellation_NotFound(t *testing.T) {
	repo := &mockReservationRepo{findByIDResult: nil}
	svc := NewBookingService(repo, nil, nil, nil, nil)
	
	_, err := svc.PreviewCancellation(context.Background(), "non-existent")
	if err == nil {
		t.Error("Expected error for not found reservation")
	}
}

func TestBookingService_PreviewCancellation_AlreadyCancelled(t *testing.T) {
	dateRange, _ := vo.NewDateRange(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 3, 23, 59, 59, 0, time.UTC))
	res := domain.Reconstitute(
		"res-1", "prop-1", "ut-1", "rp-1", "", "guest-1",
		dateRange,
		vo.NewMoney(20000, "USD"),
		domain.StatusCancelled, // Already cancelled
		"test@test.com", "RES123", time.Now(),
	)
	repo := &mockReservationRepo{findByIDResult: res}
	svc := NewBookingService(repo, nil, nil, nil, nil)
	
	_, err := svc.PreviewCancellation(context.Background(), "res-1")
	if err == nil {
		t.Error("Expected error for already cancelled reservation")
	}
}

func TestBookingService_PreviewCancellation_Success(t *testing.T) {
	dateRange, _ := vo.NewDateRange(time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 7, 3, 23, 59, 59, 0, time.UTC))
	res := domain.Reconstitute(
		"res-1", "prop-1", "ut-1", "rp-1", "", "guest-1",
		dateRange,
		vo.NewMoney(50000, "USD"),
		domain.StatusConfirmed,
		"test@test.com", "RES123", time.Now(),
	)
	repo := &mockReservationRepo{findByIDResult: res}
	pricing := &mockPricingService{calculatePenaltyResult: vo.NewMoney(10000, "USD")}
	svc := NewBookingService(repo, pricing, nil, nil, nil)
	
	penalty, err := svc.PreviewCancellation(context.Background(), "res-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if penalty != 100.00 {
		t.Errorf("Expected penalty 100.00, got: %f", penalty)
	}
}

func TestBookingService_PreviewCancellation_PricingError(t *testing.T) {
	dateRange, _ := vo.NewDateRange(time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 7, 3, 23, 59, 59, 0, time.UTC))
	res := domain.Reconstitute(
		"res-1", "prop-1", "ut-1", "rp-1", "", "guest-1",
		dateRange,
		vo.NewMoney(50000, "USD"),
		domain.StatusConfirmed,
		"test@test.com", "RES123", time.Now(),
	)
	repo := &mockReservationRepo{findByIDResult: res}
	pricing := &mockPricingService{calculatePenaltyErr: errors.New("pricing error")}
	svc := NewBookingService(repo, pricing, nil, nil, nil)
	
	_, err := svc.PreviewCancellation(context.Background(), "res-1")
	if err == nil {
		t.Error("Expected error from pricing service")
	}
}
