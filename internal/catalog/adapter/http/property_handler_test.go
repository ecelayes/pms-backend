package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestUpdatePropertyRequest_JSON(t *testing.T) {
	jsonStr := `{"name":"Updated Hotel","code":"UPD-001","type":"apartment"}`
	var req UpdatePropertyRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}
	if req.Name != "Updated Hotel" {
		t.Errorf("Expected name 'Updated Hotel', got '%s'", req.Name)
	}
	if req.Code != "UPD-001" {
		t.Errorf("Expected code 'UPD-001', got '%s'", req.Code)
	}
	if req.Type != "apartment" {
		t.Errorf("Expected type 'apartment', got '%s'", req.Type)
	}
}

func TestPropertyHandler_Create_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/properties", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &PropertyHandler{service: nil}
	_ = h.Create(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestPropertyHandler_Update_InvalidJSON(t *testing.T) {
	e := echo.New()
	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPut, "/properties/prop-123", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("prop-123")

	h := &PropertyHandler{service: nil}
	_ = h.Update(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
