package application

import (
	"sync"
	"sync/atomic"
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


type mockStreamPublisher struct {
	publishCreatedErr      error
	publishCancelledErr    error
	publishConfirmedErr    error
	ensureGroupsErr        error
	publishCreatedCalled   int32
	publishCancelledCalled int32
	publishConfirmedCalled int32
	ensureGroupsCalled     int32
	waitForPublish         bool
}

func (m *mockStreamPublisher) PublishReservationCreated(ctx context.Context, payload sharedDomain.ReservationCreatedPayload) error {
	atomic.AddInt32(&m.publishCreatedCalled, 1)
	if m.waitForPublish {
		// Block until test is done checking coverage
		time.Sleep(100 * time.Millisecond)
	}
	return m.publishCreatedErr
}

func (m *mockStreamPublisher) PublishReservationCancelled(ctx context.Context, payload sharedDomain.ReservationCancelledPayload) error {
	atomic.AddInt32(&m.publishCancelledCalled, 1)
	if m.waitForPublish {
		time.Sleep(100 * time.Millisecond)
	}
	return m.publishCancelledErr
}

func (m *mockStreamPublisher) PublishReservationConfirmed(ctx context.Context, payload sharedDomain.ReservationConfirmedPayload) error {
	atomic.AddInt32(&m.publishConfirmedCalled, 1)
	return m.publishConfirmedErr
}

func (m *mockStreamPublisher) EnsureGroups(ctx context.Context) error {
	atomic.AddInt32(&m.ensureGroupsCalled, 1)
	return m.ensureGroupsErr
}

func futureDateRange() vo.DateRange {
	now := time.Now()
	start := now.AddDate(0, 1, 0)
	end := start.AddDate(0, 0, 2)
	dr, _ := vo.NewDateRange(start, end)
	return dr
}

func TestNewBookingService(t *testing.T) {
	svc := NewBookingService(nil, nil, nil, nil, nil, nil, nil)
	if svc == nil {
		t.Error("Expected non-nil service")
	}
}

func TestNewBookingServiceWithPublisher(t *testing.T) {
	svc := NewBookingService(nil, nil, nil, nil, nil, nil, nil)
	if svc == nil {
		t.Error("Expected non-nil service")
	}
}

func TestNewBookingServiceNilPublisher(t *testing.T) {
	svc := NewBookingService(nil, nil, nil, nil, nil, nil, nil)
	if svc == nil {
		t.Error("Expected non-nil service")
	}
}

func TestBookingService_CreateReservation_InvalidDateRange(t *testing.T) {
	repo := &mockReservationRepo{}
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, nil)
	
	dr := futureDateRange()
	start := dr.End()
	end := dr.Start()
	
	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err == nil {
		t.Error("Expected error for invalid date range")
	}
}

func TestBookingService_CreateReservation_IdentityServiceError(t *testing.T) {
	repo := &mockReservationRepo{}
	identity := &mockIdentityService{findOrCreateGuestErr: errors.New("identity error")}
	svc := NewBookingService(repo, nil, nil, identity, nil, nil, nil)
	
	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()
	
	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err == nil {
		t.Error("Expected error from identity service")
	}
}

func TestBookingService_CreateReservation_TransactionFailsOnLock(t *testing.T) {
	repo := &mockReservationRepo{lockUnitTypeErr: errors.New("lock error")}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-1"}
	svc := NewBookingService(repo, nil, nil, identity, nil, nil, nil)
	
	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()
	
	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err == nil {
		t.Error("Expected error from lock")
	}
}

func TestBookingService_CreateReservation_CatalogError(t *testing.T) {
	repo := &mockReservationRepo{}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-1"}
	catalog := &mockCatalogService{getUnitTypeErr: errors.New("catalog error")}
	svc := NewBookingService(repo, nil, catalog, identity, nil, nil, nil)
	
	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()
	
	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err == nil {
		t.Error("Expected error from catalog")
	}
}

func TestBookingService_CreateReservation_NoAvailability(t *testing.T) {
	repo := &mockReservationRepo{countOverlappingCnt: 10}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-1"}
	catalog := &mockCatalogService{getUnitTypePropertyID: "prop-1", getUnitTypeBasePrice: vo.NewMoney(10000, "USD"), getUnitTypeTotalQty: 10}
	svc := NewBookingService(repo, nil, catalog, identity, nil, nil, nil)
	
	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()
	
	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err != ErrNoAvailability {
		t.Errorf("Expected ErrNoAvailability, got: %v", err)
	}
}

func TestBookingService_CreateReservation_PricingError(t *testing.T) {
	repo := &mockReservationRepo{countOverlappingCnt: 0}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-1"}
	catalog := &mockCatalogService{getUnitTypePropertyID: "prop-1", getUnitTypeBasePrice: vo.NewMoney(10000, "USD"), getUnitTypeTotalQty: 10}
	pricing := &mockPricingService{calculateStayPriceErr: errors.New("pricing error")}
	svc := NewBookingService(repo, pricing, catalog, identity, nil, nil, nil)
	
	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()
	
	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err == nil {
		t.Error("Expected error from pricing")
	}
}

func TestBookingService_CreateReservation_Success(t *testing.T) {
	repo := &mockReservationRepo{}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-1"}
	catalog := &mockCatalogService{getUnitTypePropertyID: "prop-1", getUnitTypeBasePrice: vo.NewMoney(10000, "USD"), getUnitTypeTotalQty: 10}
	pricing := &mockPricingService{calculateStayPriceResult: vo.NewMoney(10000, "USD")}
	svc := NewBookingService(repo, pricing, catalog, identity, nil, nil, nil)
	
	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()
	
	code, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
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
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, nil)
	
	res, err := svc.GetReservationByCode(context.Background(), "RES-123")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res != expectedRes {
		t.Error("Expected reservation")
	}
}


func TestBookingService_CancelReservation_NotFound(t *testing.T) {
	repo := &mockReservationRepo{findByIDResult: nil}
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, nil)
	
	err := svc.CancelReservation(context.Background(), "res-1")
	if err == nil {
		t.Error("Expected error for not found")
	}
}

func TestBookingService_CancelReservation_Success(t *testing.T) {
	res := &domain.Reservation{}
	repo := &mockReservationRepo{findByIDResult: res}
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, nil)
	
	err := svc.CancelReservation(context.Background(), "res-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestBookingService_CancelReservation_UpdateError(t *testing.T) {
	res := &domain.Reservation{}
	repo := &mockReservationRepo{findByIDResult: res, updateErr: errors.New("update error")}
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, nil)
	
	err := svc.CancelReservation(context.Background(), "res-1")
	if err == nil {
		t.Error("Expected error from update")
	}
}

func TestBookingService_PreviewCancellation_NotFound(t *testing.T) {
	repo := &mockReservationRepo{findByIDResult: nil}
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, nil)
	
	_, err := svc.PreviewCancellation(context.Background(), "res-1")
	if err == nil {
		t.Error("Expected error for not found")
	}
}

func TestBookingService_PreviewCancellation_AlreadyCancelled(t *testing.T) {
	res := &domain.Reservation{}
	res.Cancel()
	repo := &mockReservationRepo{findByIDResult: res}
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, nil)
	
	_, err := svc.PreviewCancellation(context.Background(), "res-1")
	if err == nil {
		t.Error("Expected error for already cancelled")
	}
}

func TestBookingService_PreviewCancellation_Success(t *testing.T) {
	res := &domain.Reservation{}
	repo := &mockReservationRepo{findByIDResult: res}
	pricing := &mockPricingService{calculatePenaltyResult: vo.NewMoney(5000, "USD")}
	svc := NewBookingService(repo, pricing, nil, nil, nil, nil, nil)
	
	penalty, err := svc.PreviewCancellation(context.Background(), "res-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if penalty != 50.0 {
		t.Errorf("Expected 50.0, got: %f", penalty)
	}
}

func TestBookingService_PreviewCancellation_PricingError(t *testing.T) {
	res := &domain.Reservation{}
	repo := &mockReservationRepo{findByIDResult: res}
	pricing := &mockPricingService{calculatePenaltyErr: errors.New("pricing error")}
	svc := NewBookingService(repo, pricing, nil, nil, nil, nil, nil)
	
	_, err := svc.PreviewCancellation(context.Background(), "res-1")
	if err == nil {
		t.Error("Expected error from pricing service")
	}
}

func TestBookingService_CancelReservation_WithPublisher(t *testing.T) {
	res := &domain.Reservation{}
	repo := &mockReservationRepo{findByIDResult: res}
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, nil)

	err := svc.CancelReservation(context.Background(), "res-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}



func TestBookingService_CreateReservation_ConfirmError(t *testing.T) {
	repo := &mockReservationRepo{}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-1"}
	catalog := &mockCatalogService{getUnitTypePropertyID: "prop-1", getUnitTypeBasePrice: vo.NewMoney(10000, "USD"), getUnitTypeTotalQty: 10}
	pricing := &mockPricingService{calculateStayPriceResult: vo.NewMoney(10000, "USD")}
	svc := NewBookingService(repo, pricing, catalog, identity, nil, nil, nil)

	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()

	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 2, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestBookingService_CreateReservation_SaveError(t *testing.T) {
	repo := &mockReservationRepo{saveErr: errors.New("save error")}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-1"}
	catalog := &mockCatalogService{getUnitTypePropertyID: "prop-1", getUnitTypeBasePrice: vo.NewMoney(10000, "USD"), getUnitTypeTotalQty: 10}
	pricing := &mockPricingService{calculateStayPriceResult: vo.NewMoney(10000, "USD")}
	svc := NewBookingService(repo, pricing, catalog, identity, nil, nil, nil)

	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()

	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 2, 0)
	if err == nil {
		t.Error("Expected error from Save")
	}
}

func TestBookingService_CreateReservation_WithPublisherSuccess(t *testing.T) {
	dr := futureDateRange()
	res, _ := domain.NewReservation("prop-1", "ut-1", "rp-1", "guest-1", dr, vo.NewMoney(10000, "USD"), "test@test.com")
	repo := &mockReservationRepo{getByCodeResult: res}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-1"}
	catalog := &mockCatalogService{getUnitTypePropertyID: "prop-1", getUnitTypeBasePrice: vo.NewMoney(10000, "USD"), getUnitTypeTotalQty: 10}
	pricing := &mockPricingService{calculateStayPriceResult: vo.NewMoney(10000, "USD")}
	publisher := &mockStreamPublisher{waitForPublish: true}
	svc := NewBookingService(repo, pricing, catalog, identity, nil, nil, publisher)

	start := dr.Start()
	end := dr.End()

	code, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 2, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if code == "" {
		t.Error("Expected non-empty reservation code")
	}
	time.Sleep(200 * time.Millisecond) // Wait for goroutine
}

func TestBookingService_CreateReservation_WithPublisherGetByCodeError(t *testing.T) {
	repo := &mockReservationRepo{getByCodeErr: errors.New("get by code error")}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-1"}
	catalog := &mockCatalogService{getUnitTypePropertyID: "prop-1", getUnitTypeBasePrice: vo.NewMoney(10000, "USD"), getUnitTypeTotalQty: 10}
	pricing := &mockPricingService{calculateStayPriceResult: vo.NewMoney(10000, "USD")}
	publisher := &mockStreamPublisher{waitForPublish: true}
	svc := NewBookingService(repo, pricing, catalog, identity, nil, nil, publisher)

	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()

	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 2, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	time.Sleep(200 * time.Millisecond)
}


func TestBookingService_GetReservationByCode_NotFound(t *testing.T) {
	repo := &mockReservationRepo{getByCodeResult: nil, getByCodeErr: errors.New("not found")}
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, nil)

	_, err := svc.GetReservationByCode(context.Background(), "NON-EXISTENT")
	if err == nil {
		t.Error("Expected error for not found")
	}
}

func TestBookingService_CancelReservation_PublisherGetByCodeError(t *testing.T) {
	res := &domain.Reservation{}
	repo := &mockReservationRepo{findByIDResult: res, getByCodeErr: errors.New("get by code error")}
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, nil)

	err := svc.CancelReservation(context.Background(), "res-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestBookingService_PreviewCancellation_FindByIDError(t *testing.T) {
	repo := &mockReservationRepo{findByIDErr: errors.New("find error")}
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, nil)

	_, err := svc.PreviewCancellation(context.Background(), "res-1")
	if err == nil {
		t.Error("Expected error from FindByID")
	}
}

func TestBookingService_CancelReservation_FindByIDError(t *testing.T) {
	repo := &mockReservationRepo{findByIDErr: errors.New("find error")}
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, nil)

	err := svc.CancelReservation(context.Background(), "res-1")
	if err == nil {
		t.Error("Expected error from FindByID")
	}
}

func TestBookingService_GetReservationByCode_RepoError(t *testing.T) {
	repo := &mockReservationRepo{getByCodeErr: errors.New("get error")}
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, nil)

	_, err := svc.GetReservationByCode(context.Background(), "RES-123")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestBookingService_CancelReservation_NilReservation(t *testing.T) {
	repo := &mockReservationRepo{findByIDResult: nil}
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, nil)

	err := svc.CancelReservation(context.Background(), "non-existent")
	if err == nil {
		t.Error("Expected error for nil reservation")
	}
}

func TestBookingService_PreviewCancellation_NilReservation(t *testing.T) {
	repo := &mockReservationRepo{findByIDResult: nil}
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, nil)

	_, err := svc.PreviewCancellation(context.Background(), "non-existent")
	if err == nil {
		t.Error("Expected error for nil reservation")
	}
}

func TestBookingService_CreateReservation_CountOverlappingError(t *testing.T) {
	repo := &mockReservationRepo{countOverlappingErr: errors.New("count error")}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-1"}
	catalog := &mockCatalogService{getUnitTypePropertyID: "prop-1", getUnitTypeBasePrice: vo.NewMoney(10000, "USD"), getUnitTypeTotalQty: 10}
	svc := NewBookingService(repo, nil, catalog, identity, nil, nil, nil)

	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()

	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err == nil {
		t.Error("Expected error from CountOverlapping")
	}
}

func TestBookingService_CreateReservation_NewReservationError(t *testing.T) {
	repo := &mockReservationRepo{countOverlappingCnt: 0}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-1"}
	catalog := &mockCatalogService{getUnitTypePropertyID: "prop-1", getUnitTypeBasePrice: vo.NewMoney(10000, "USD"), getUnitTypeTotalQty: 10}
	pricing := &mockPricingService{calculateStayPriceResult: vo.NewMoney(10000, "USD")}
	svc := NewBookingService(repo, pricing, catalog, identity, nil, nil, nil)

	// Use past dates to trigger NewReservation error
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err == nil {
		t.Error("Expected error from NewReservation (past dates)")
	}
}

func TestBookingService_CreateReservation_PublisherPublishError(t *testing.T) {
	dr := futureDateRange()
	res, _ := domain.NewReservation("prop-1", "ut-1", "rp-1", "guest-1", dr, vo.NewMoney(10000, "USD"), "test@test.com")
	repo := &mockReservationRepo{getByCodeResult: res}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-1"}
	catalog := &mockCatalogService{getUnitTypePropertyID: "prop-1", getUnitTypeBasePrice: vo.NewMoney(10000, "USD"), getUnitTypeTotalQty: 10}
	pricing := &mockPricingService{calculateStayPriceResult: vo.NewMoney(10000, "USD")}
	publisher := &mockStreamPublisher{publishCreatedErr: errors.New("publish error"), waitForPublish: true}
	svc := NewBookingService(repo, pricing, catalog, identity, nil, nil, publisher)

	start := dr.Start()
	end := dr.End()

	code, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if code == "" {
		t.Error("Expected non-empty code")
	}
	time.Sleep(200 * time.Millisecond)
}

func TestBookingService_CreateReservation_PublisherGetByCodeNil(t *testing.T) {
	repo := &mockReservationRepo{getByCodeResult: nil, getByCodeErr: nil}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-1"}
	catalog := &mockCatalogService{getUnitTypePropertyID: "prop-1", getUnitTypeBasePrice: vo.NewMoney(10000, "USD"), getUnitTypeTotalQty: 10}
	pricing := &mockPricingService{calculateStayPriceResult: vo.NewMoney(10000, "USD")}
	svc := NewBookingService(repo, pricing, catalog, identity, nil, nil, nil)

	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()

	code, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if code == "" {
		t.Error("Expected non-empty code")
	}
}

func TestBookingService_CancelReservation_PublisherError(t *testing.T) {
	dr := futureDateRange()
	res, _ := domain.NewReservation("prop-1", "ut-1", "rp-1", "guest-1", dr, vo.NewMoney(10000, "USD"), "test@test.com")
	res.Confirm()
	repo := &mockReservationRepo{findByIDResult: res}
	publisher := &mockStreamPublisher{publishCancelledErr: errors.New("publish cancelled error"), waitForPublish: true}
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, publisher)

	err := svc.CancelReservation(context.Background(), "res-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	time.Sleep(200 * time.Millisecond)
}



func TestBookingService_CancelReservation_WithPublisherSuccess(t *testing.T) {
	dr := futureDateRange()
	res, _ := domain.NewReservation("prop-1", "ut-1", "rp-1", "guest-1", dr, vo.NewMoney(10000, "USD"), "test@test.com")
	res.Confirm()
	repo := &mockReservationRepo{findByIDResult: res}
	publisher := &mockStreamPublisher{waitForPublish: true}
	svc := NewBookingService(repo, nil, nil, nil, nil, nil, publisher)

	err := svc.CancelReservation(context.Background(), "res-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	time.Sleep(200 * time.Millisecond)
}


// mockReservationFactory lets tests return a pre-built reservation,
// including one that has already been confirmed (to exercise the
// ErrReservationAlreadyConfirmed path).
type mockReservationFactory struct {
	result *domain.Reservation
	err    error
}

func (m *mockReservationFactory) Create(
	propertyID, unitTypeID, ratePlanID, guestID string,
	dateRange vo.DateRange,
	price vo.Money,
	guestEmail string,
) (*domain.Reservation, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func TestBookingService_CreateReservation_AlreadyConfirmedReturnsError(t *testing.T) {
	// Build a reservation and pre-confirm it. Confirm() will return
	// ErrReservationAlreadyConfirmed when called again.
	preConfirmed, err := domain.NewReservation(
		"prop-1", "ut-1", "rp-1", "guest-1",
		futureDateRange(), vo.NewMoney(10000, "USD"), "test@test.com",
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := preConfirmed.Confirm(); err != nil {
		t.Fatalf("setup: %v", err)
	}

	repo := &mockReservationRepo{}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-1"}
	catalog := &mockCatalogService{getUnitTypePropertyID: "prop-1", getUnitTypeBasePrice: vo.NewMoney(10000, "USD"), getUnitTypeTotalQty: 10}
	pricing := &mockPricingService{calculateStayPriceResult: vo.NewMoney(10000, "USD")}
	svc := NewBookingService(repo, pricing, catalog, identity, nil, nil, nil)
	svc.factory = &mockReservationFactory{result: preConfirmed}

	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()

	_, err = svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if !errors.Is(err, domain.ErrReservationAlreadyConfirmed) {
		t.Errorf("expected ErrReservationAlreadyConfirmed, got: %v", err)
	}
}

func TestBookingService_CreateReservation_FactoryError(t *testing.T) {
	repo := &mockReservationRepo{}
	identity := &mockIdentityService{findOrCreateGuestID: "guest-1"}
	catalog := &mockCatalogService{getUnitTypePropertyID: "prop-1", getUnitTypeBasePrice: vo.NewMoney(10000, "USD"), getUnitTypeTotalQty: 10}
	pricing := &mockPricingService{calculateStayPriceResult: vo.NewMoney(10000, "USD")}
	svc := NewBookingService(repo, pricing, catalog, identity, nil, nil, nil)
	svc.factory = &mockReservationFactory{err: errors.New("factory boom")}

	dr := futureDateRange()
	start := dr.Start()
	end := dr.End()

	_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", start, end, "test@test.com", "John", "Doe", "123", 1, 0)
	if err == nil {
		t.Error("expected factory error, got nil")
	}
}


// TestBookingService_CreateReservation_ConcurrentInventoryRace verifies that
// when multiple goroutines attempt to create a reservation against the same
// unit type with limited inventory, the row-level lock + count check ensures
// only (totalQty) reservations succeed.
func TestBookingService_CreateReservation_ConcurrentInventoryRace(t *testing.T) {
	const totalQty = 2
	const goroutines = 8

	var successCount atomic.Int32
	var conflictCount atomic.Int32
	var otherCount atomic.Int32

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Shared counter simulates the inventory check happening inside a real DB
	// transaction with row locks. The first `totalQty` callers succeed; the rest
	// receive ErrNoAvailability.
	var inv atomic.Int32
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			repo := &mockReservationRepo{}
			identity := &mockIdentityService{findOrCreateGuestID: "guest-1"}
			catalog := &mockCatalogService{
				getUnitTypePropertyID: "prop-1",
				getUnitTypeBasePrice:   vo.NewMoney(10000, "USD"),
				getUnitTypeTotalQty:   totalQty,
			}
			pricing := &mockPricingService{calculateStayPriceResult: vo.NewMoney(10000, "USD")}

			repo.runInTxFunc = func(ctx context.Context, fn func(context.Context) error) error {
				if inv.Add(1) > int32(totalQty) {
					return ErrNoAvailability
				}
				return fn(ctx)
			}

			svc := NewBookingService(repo, pricing, catalog, identity, nil, nil, nil)
			dr := futureDateRange()
			_, err := svc.CreateReservation(context.Background(), "ut-1", "rp-1", dr.Start(), dr.End(), "test@test.com", "John", "Doe", "123", 1, 0)
			switch {
			case err == nil:
				successCount.Add(1)
			case errors.Is(err, ErrNoAvailability):
				conflictCount.Add(1)
			default:
				otherCount.Add(1)
			}
		}()
	}
	wg.Wait()

	if got := successCount.Load(); got > int32(totalQty) {
		t.Errorf("expected at most %d successes, got %d", totalQty, got)
	}
	if got := successCount.Load(); got < 1 {
		t.Error("expected at least one success")
	}
	if otherCount.Load() != 0 {
		t.Errorf("unexpected non-availability errors: %d", otherCount.Load())
	}
}
