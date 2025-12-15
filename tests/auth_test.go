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
	s.db.Exec(ctx, `INSERT INTO users (id, email, password, salt, role, first_name, last_name, phone, created_at, updated_at) VALUES ($1, $2, $3, $4, 'user', 'Test', 'User', '12345', NOW(), NOW())`, userID.String(), email, hash, salt)
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

func (s *AuthSuite) TestPasswordResetFlow() {
	ctx := context.Background()
	userID, _ := uuid.NewV7()
	email := "reset@test.com"
	pass := "oldpass"
	hash, _ := auth.HashPassword(pass)
	salt, _ := auth.GenerateRandomSalt()

	_, err := s.db.Exec(ctx, `INSERT INTO users (id, email, password, salt, role, first_name, last_name, phone, created_at, updated_at) VALUES ($1, $2, $3, $4, 'user', 'Test', 'User', '12345', NOW(), NOW())`, userID.String(), email, hash, salt)
	s.Require().NoError(err)

	resReq := s.MakeRequest("POST", "/api/v1/auth/forgot-password", map[string]string{"email": email}, "")
	s.Equal(http.StatusOK, resReq.Code)

	var currentSalt string
	err = s.db.QueryRow(ctx, "SELECT salt FROM users WHERE id=$1", userID).Scan(&currentSalt)
	s.Require().NoError(err)

	resetToken, err := auth.GenerateResetToken(userID.String(), currentSalt)
	s.Require().NoError(err)

	newPass := "newpass123"
	resReset := s.MakeRequest("POST", "/api/v1/auth/reset-password", map[string]string{
		"token": resetToken,
		"new_password": newPass,
	}, "")
	s.Equal(http.StatusOK, resReset.Code, "El reset debería funcionar con un token válido")

	resLogin := s.MakeRequest("POST", "/api/v1/auth/login", map[string]string{
		"email": email,
		"password": newPass,
	}, "")
	s.Equal(http.StatusOK, resLogin.Code, "El login con la nueva contraseña debería funcionar")

	resOldLogin := s.MakeRequest("POST", "/api/v1/auth/login", map[string]string{
		"email": email,
		"password": pass,
	}, "")
	s.Equal(http.StatusUnauthorized, resOldLogin.Code)

	resReplay := s.MakeRequest("POST", "/api/v1/auth/reset-password", map[string]string{
		"token": resetToken,
		"new_password": "hacker_attempt",
	}, "")
	s.Equal(http.StatusUnauthorized, resReplay.Code, "El token usado no debería servir una segunda vez")
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthSuite))
}
