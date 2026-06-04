package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCreateReservationRequest_JSON(t *testing.T) {
	jsonStr := `{
		"unit_type_id":"ut-123",
		"start":"2024-06-01",
		"end":"2024-06-03",
		"guest_email":"test@example.com",
		"guest_first_name":"John",
		"guest_last_name":"Doe",
		"guest_phone":"+1234567890",
		"rate_plan_id":"rp-456",
		"adults":2,
		"children":1
	}`
	var req CreateReservationRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}
	if req.UnitTypeID != "ut-123" {
		t.Errorf("Expected unit_type_id 'ut-123', got '%s'", req.UnitTypeID)
	}
	if req.Start != "2024-06-01" {
		t.Errorf("Expected start '2024-06-01', got '%s'", req.Start)
	}
	if req.End != "2024-06-03" {
		t.Errorf("Expected end '2024-06-03', got '%s'", req.End)
	}
	if req.GuestEmail != "test@example.com" {
		t.Errorf("Expected guest_email 'test@example.com', got '%s'", req.GuestEmail)
	}
	if req.Adults != 2 {
		t.Errorf("Expected adults 2, got %d", req.Adults)
	}
	if req.Children != 1 {
		t.Errorf("Expected children 1, got %d", req.Children)
	}
}

func TestReservationHandler_Create_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &ReservationHandler{service: nil}
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestReservationHandler_Create_MissingUnitTypeID(t *testing.T) {
	e := echo.New()
	reqBody := `{"start":"2024-06-01","end":"2024-06-03","guest_email":"test@example.com","rate_plan_id":"rp-456"}`
	req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &ReservationHandler{service: nil}
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestReservationHandler_Create_MissingGuestEmail(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-123","start":"2024-06-01","end":"2024-06-03","rate_plan_id":"rp-456"}`
	req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &ReservationHandler{service: nil}
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestReservationHandler_Create_MissingRatePlanID(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-123","start":"2024-06-01","end":"2024-06-03","guest_email":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &ReservationHandler{service: nil}
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestReservationHandler_Create_InvalidStartDate(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-123","start":"invalid-date","end":"2024-06-03","guest_email":"test@example.com","rate_plan_id":"rp-456"}`
	req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &ReservationHandler{service: nil}
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestReservationHandler_Create_InvalidEndDate(t *testing.T) {
	e := echo.New()
	reqBody := `{"unit_type_id":"ut-123","start":"2024-06-01","end":"invalid-date","guest_email":"test@example.com","rate_plan_id":"rp-456"}`
	req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &ReservationHandler{service: nil}
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestReservationHandler_Create_ValidJSONParsing(t *testing.T) {
	e := echo.New()
	// Use negative price to trigger validation error before service call
	reqBody := `{"unit_type_id":"ut-123","start":"2024-06-01","end":"invalid-date","guest_email":"test@example.com","rate_plan_id":"rp-456"}`
	req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &ReservationHandler{service: nil}
	_ = h.Create(c)

	// Should fail on invalid end date
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
