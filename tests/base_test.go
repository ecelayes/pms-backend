package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"github.com/ecelayes/pms-backend/internal/bootstrap"
	"github.com/ecelayes/pms-backend/pkg/auth"
)

type BaseSuite struct {
	suite.Suite
	echo *echo.Echo
	db   *pgxpool.Pool
}

func (s *BaseSuite) SetupSuite() {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" { s.T().Fatal("TEST_DATABASE_URL is not set") }
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil { s.T().Fatal(err) }
	s.db = pool
	s.echo = bootstrap.NewApp(pool)
}

func (s *BaseSuite) TearDownSuite() { s.db.Close() }

func (s *BaseSuite) SetupTest() {
	tables := []string{"reservations", "price_rules", "room_types", "hotels", "organization_members", "users", "organizations", "guests"}
	for _, table := range tables {
		s.db.Exec(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
	}
}

func (s *BaseSuite) MakeRequest(method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	var bodyReader *bytes.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(jsonBody)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}
	req := httptest.NewRequest(method, url, bodyReader)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if token != "" {
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	return rec
}

func (s *BaseSuite) GetAdminTokenAndOrg() (string, string) {
	ctx := context.Background()
	
	orgID, _ := uuid.NewV7()
	userID, _ := uuid.NewV7()
	memberID, _ := uuid.NewV7()
	
	email := "admin@test.com"
	password := "password123"
	hashedPass, _ := auth.HashPassword(password)
	salt, _ := auth.GenerateRandomSalt()

	_, err := s.db.Exec(ctx, `INSERT INTO organizations (id, name, code, created_at, updated_at) VALUES ($1, 'Test Corp', 'TEST', NOW(), NOW())`, orgID.String())
	if err != nil { s.T().Fatal(err) }

	_, err = s.db.Exec(ctx, `INSERT INTO users (id, email, password, salt, role, created_at, updated_at) VALUES ($1, $2, $3, $4, 'user', NOW(), NOW())`, userID.String(), email, hashedPass, salt)
	if err != nil { s.T().Fatal(err) }

	_, err = s.db.Exec(ctx, `INSERT INTO organization_members (id, organization_id, user_id, role, created_at, updated_at) VALUES ($1, $2, $3, 'owner', NOW(), NOW())`, memberID.String(), orgID.String(), userID.String())
	if err != nil { s.T().Fatal(err) }

	resLogin := s.MakeRequest("POST", "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, "")
	
	var dataLogin map[string]string
	json.Unmarshal(resLogin.Body.Bytes(), &dataLogin)
	
	return dataLogin["token"], orgID.String()
}

func (s *BaseSuite) GetOrgIDByOwnerEmail(email string) string {
	var orgID string
	err := s.db.QueryRow(context.Background(), `SELECT organization_id FROM organization_members om JOIN users u ON om.user_id = u.id WHERE u.email = $1`, email).Scan(&orgID)
	if err != nil { s.T().Fatal(err) }
	return orgID
}
