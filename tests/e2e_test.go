package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/ecelayes/pms-backend/internal/bootstrap"
)

type E2ESuite struct {
	suite.Suite
	db      *pgxpool.Pool
	handler http.Handler
	
	token      string
	hotelID    string
	roomTypeID string
	resCode    string
}

func (s *E2ESuite) SetupSuite() {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/hotel_pms_test?sslmode=disable"
	}

	var err error
	s.db, err = pgxpool.New(context.Background(), dsn)
	s.Require().NoError(err)

	e := bootstrap.NewApp(s.db)
	s.handler = e
}

func (s *E2ESuite) SetupTest() {
	_, err := s.db.Exec(context.Background(), "TRUNCATE reservations, price_rules, room_types, hotels, users CASCADE")
	s.Require().NoError(err)
}

func (s *E2ESuite) TearDownSuite() {
	s.db.Close()
}

func (s *E2ESuite) makeRequest(method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	var bodyReader *bytes.Buffer
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonBytes)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	s.handler.ServeHTTP(rec, req)
	return rec
}

func (s *E2ESuite) TestFullFlow() {
	s.Run("1. Register Admin", func() {
		res := s.makeRequest("POST", "/api/v1/auth/register", map[string]string{
			"email":    "admin@test.com",
			"password": "Password123!",
		}, "")
		s.Equal(http.StatusCreated, res.Code)
	})

	s.Run("2. Login Admin", func() {
		res := s.makeRequest("POST", "/api/v1/auth/login", map[string]string{
			"email":    "admin@test.com",
			"password": "Password123!",
		}, "")
		s.Equal(http.StatusOK, res.Code)

		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.token = data["token"]
		s.NotEmpty(s.token)
	})

	s.Run("3. Create Hotel", func() {
		res := s.makeRequest("POST", "/api/v1/hotels", map[string]string{
			"name": "Test Hotel",
			"code": "TST",
		}, s.token)
		s.Equal(http.StatusCreated, res.Code)

		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.hotelID = data["hotel_id"]
		s.NotEmpty(s.hotelID)
	})

	s.Run("4. Create Room Type", func() {
		res := s.makeRequest("POST", "/api/v1/room-types", map[string]interface{}{
			"hotel_id":       s.hotelID,
			"name":           "Suite Test",
			"code":           "SUI",
			"total_quantity": 5,
		}, s.token)
		s.Equal(http.StatusCreated, res.Code)

		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.roomTypeID = data["room_type_id"]
		s.NotEmpty(s.roomTypeID)
	})

	s.Run("5. Create Pricing Rule", func() {
		res := s.makeRequest("POST", "/api/v1/pricing/rules", map[string]interface{}{
			"room_type_id": s.roomTypeID,
			"start":        "2025-01-01",
			"end":          "2025-01-05",
			"price":        100.0,
			"priority":     10,
		}, s.token)
		s.Equal(http.StatusCreated, res.Code)
	})

	s.Run("6. Create Reservation", func() {
		res := s.makeRequest("POST", "/api/v1/reservations", map[string]interface{}{
			"room_type_id": s.roomTypeID,
			"guest_email":  "guest@test.com",
			"start":        "2025-01-01",
			"end":          "2025-01-05",
		}, "")
		
		s.Equal(http.StatusCreated, res.Code)

		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.resCode = data["reservation_code"]
		
		s.Contains(s.resCode, "TST-SUI-")
	})

	s.Run("7. Check Availability", func() {
		res := s.makeRequest("GET", "/api/v1/availability?start=2025-01-01&end=2025-01-05", nil, "")
		s.Equal(http.StatusOK, res.Code)
		
		type AvailResponse struct {
			Data []map[string]interface{} `json:"data"`
		}
		var resp AvailResponse
		json.Unmarshal(res.Body.Bytes(), &resp)

		found := false
		for _, item := range resp.Data {
			if item["room_type_id"] == s.roomTypeID {
				qty := item["available_qty"].(float64)
				s.Equal(4.0, qty)
				found = true
			}
		}
		s.True(found)
	})

	s.Run("8. Cancel Reservation", func() {
		resGet := s.makeRequest("GET", "/api/v1/reservations/"+s.resCode, nil, "")
		s.Equal(http.StatusOK, resGet.Code)
		var resData map[string]interface{}
		json.Unmarshal(resGet.Body.Bytes(), &resData)
		id := resData["id"].(string)

		resCancel := s.makeRequest("POST", "/api/v1/reservations/"+id+"/cancel", nil, "")
		s.Equal(http.StatusOK, resCancel.Code)
	})
}

func TestE2E(t *testing.T) {
	suite.Run(t, new(E2ESuite))
}
