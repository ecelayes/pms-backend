package http

import (
	"errors"
	"time"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ecelayes/pms-backend/internal/booking/application"
	"github.com/ecelayes/pms-backend/internal/booking/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"github.com/labstack/echo/v4"
)

func TestReservationHandler_Create_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid}`
	req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewReservationHandler(&mockBookingService{})
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestReservationHandler_Create_MissingFields(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"","guest_email":"a@b.com","rate_plan_id":"rp-1"}`
	req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewReservationHandler(&mockBookingService{})
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestReservationHandler_Create_InvalidStartDate(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-1","guest_email":"a@b.com","rate_plan_id":"rp-1","start":"bad","end":"2025-12-31"}`
	req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewReservationHandler(&mockBookingService{})
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestReservationHandler_Create_InvalidEndDate(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-1","guest_email":"a@b.com","rate_plan_id":"rp-1","start":"2025-06-01","end":"bad"}`
	req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewReservationHandler(&mockBookingService{})
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestReservationHandler_Create_Success(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-1","guest_email":"a@b.com","rate_plan_id":"rp-1","start":"2027-06-01","end":"2027-06-10","adults":2,"children":0}`
	req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockBookingService{createResult: "RES-001"}
	h := NewReservationHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d", rec.Code)
	}
}

func TestReservationHandler_Create_InvalidDates(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-1","guest_email":"a@b.com","rate_plan_id":"rp-1","start":"2027-06-01","end":"2027-06-10","adults":2,"children":0}`
	req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockBookingService{createErr: domain.ErrInvalidReservationDates}
	h := NewReservationHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestReservationHandler_Create_InvalidDateRange(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-1","guest_email":"a@b.com","rate_plan_id":"rp-1","start":"2027-06-01","end":"2027-06-10","adults":2,"children":0}`
	req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockBookingService{createErr: vo.ErrinvalidDateRange}
	h := NewReservationHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestReservationHandler_Create_NoAvailability(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-1","guest_email":"a@b.com","rate_plan_id":"rp-1","start":"2027-06-01","end":"2027-06-10","adults":2,"children":0}`
	req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockBookingService{createErr: application.ErrNoAvailability}
	h := NewReservationHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusConflict {
		t.Errorf("Expected 409, got %d", rec.Code)
	}
}

func TestReservationHandler_Create_Internal(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-1","guest_email":"a@b.com","rate_plan_id":"rp-1","start":"2027-06-01","end":"2027-06-10","adults":2,"children":0}`
	req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := &mockBookingService{createErr: errors.New("db error")}
	h := NewReservationHandler(svc)
	_ = h.Create(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestReservationHandler_GetByCode_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/reservations/RES-001", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("code")
	c.SetParamValues("RES-001")

	svc := &mockBookingService{getResult: nil}
	h := NewReservationHandler(svc)
	_ = h.GetByCode(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestReservationHandler_GetByCode_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/reservations/RES-001", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("code")
	c.SetParamValues("RES-001")

	start := time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 6, 10, 0, 0, 0, 0, time.UTC)
	dr, _ := vo.NewDateRange(start, end)
	res, _ := domain.NewReservation("p-1", "ut-1", "rp-1", "guest-1", dr, vo.NewMoney(10000, "USD"), "a@b.com")
	svc := &mockBookingService{getResult: res}
	h := NewReservationHandler(svc)
	_ = h.GetByCode(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestReservationHandler_GetByCode_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/reservations/RES-001", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("code")
	c.SetParamValues("RES-001")

	svc := &mockBookingService{getErr: errors.New("db error")}
	h := NewReservationHandler(svc)
	_ = h.GetByCode(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestReservationHandler_PreviewCancel_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/reservations/res-1/cancel/preview", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("res-1")

	svc := &mockBookingService{previewAmt: 50.0}
	h := NewReservationHandler(svc)
	_ = h.PreviewCancel(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestReservationHandler_PreviewCancel_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/reservations/res-1/cancel/preview", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("res-1")

	svc := &mockBookingService{previewErr: errors.New("not found")}
	h := NewReservationHandler(svc)
	_ = h.PreviewCancel(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestReservationHandler_Cancel_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/reservations/res-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("res-1")

	svc := &mockBookingService{}
	h := NewReservationHandler(svc)
	_ = h.Cancel(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestReservationHandler_Cancel_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/reservations/res-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("res-1")

	svc := &mockBookingService{cancelErr: errors.New("not found")}
	h := NewReservationHandler(svc)
	_ = h.Cancel(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestReservationHandler_Delete(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/reservations/res-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewReservationHandler(&mockBookingService{})
	_ = h.Delete(c)

	if rec.Code != http.StatusNotImplemented {
		t.Errorf("Expected 501, got %d", rec.Code)
	}
}

func TestReservationHandler_NewReservationHandler(t *testing.T) {
	svc := &mockBookingService{}
	h := NewReservationHandler(svc)
	if h == nil || h.service == nil {
		t.Fatal("handler or service nil")
	}
}
