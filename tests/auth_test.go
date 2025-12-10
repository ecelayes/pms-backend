package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/ecelayes/pms-backend/pkg/auth"
)

type AuthSuite struct {
	BaseSuite
}

func (s *AuthSuite) TestLogin() {
	ctx := context.Background()
	orgID, _ := uuid.NewV7()
	userID, _ := uuid.NewV7()
	memberID, _ := uuid.NewV7()
	
	email := "login@test.com"
	pass := "pass123"
	hash, _ := auth.HashPassword(pass)
	salt, _ := auth.GenerateRandomSalt()

	s.db.Exec(ctx, `INSERT INTO organizations (id, name, code, created_at, updated_at) VALUES ($1, 'Auth Corp', 'AUTH', NOW(), NOW())`, orgID.String())
	s.db.Exec(ctx, `INSERT INTO users (id, email, password, salt, role, created_at, updated_at) VALUES ($1, $2, $3, $4, 'user', NOW(), NOW())`, userID.String(), email, hash, salt)
	s.db.Exec(ctx, `INSERT INTO organization_members (id, organization_id, user_id, role, created_at, updated_at) VALUES ($1, $2, $3, 'staff', NOW(), NOW())`, memberID.String(), orgID.String(), userID.String())

	res := s.MakeRequest("POST", "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": "pass123",
	}, "")
	s.Equal(http.StatusOK, res.Code)
	s.Contains(res.Body.String(), "token")

	resFail := s.MakeRequest("POST", "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": "wrong",
	}, "")
	s.Equal(http.StatusUnauthorized, resFail.Code)
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthSuite))
}
