package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/ecelayes/pms-backend/internal/bootstrap"
)

type BaseSuite struct {
	suite.Suite
	db      *pgxpool.Pool
	handler http.Handler
}

func (s *BaseSuite) SetupSuite() {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/hotel_pms_test?sslmode=disable"
	}

	var err error
	s.db, err = pgxpool.New(context.Background(), dsn)
	s.Require().NoError(err)

	s.handler = bootstrap.NewApp(s.db)
}

func (s *BaseSuite) SetupTest() {
	_, err := s.db.Exec(context.Background(), "TRUNCATE reservations, price_rules, room_types, hotels, users CASCADE")
	s.Require().NoError(err)
}

func (s *BaseSuite) TearDownSuite() {
	s.db.Close()
}

func (s *BaseSuite) MakeRequest(method, url string, body interface{}, token string) *httptest.ResponseRecorder {
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

func (s *BaseSuite) GetAdminToken() string {
	s.MakeRequest("POST", "/api/v1/auth/register", map[string]string{
		"email":    "admin@setup.com",
		"password": "Password123!",
	}, "")

	res := s.MakeRequest("POST", "/api/v1/auth/login", map[string]string{
		"email":    "admin@setup.com",
		"password": "Password123!",
	}, "")
	
	s.Require().Equal(http.StatusOK, res.Code)
	
	var data map[string]string
	json.Unmarshal(res.Body.Bytes(), &data)
	return data["token"]
}
