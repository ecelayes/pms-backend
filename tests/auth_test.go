package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AuthSuite struct {
	BaseSuite
}

func (s *AuthSuite) TestRegister() {
	res := s.MakeRequest("POST", "/api/v1/auth/register", map[string]string{
		"email":    "newuser@test.com",
		"password": "Password123!",
	}, "")
	s.Equal(http.StatusCreated, res.Code)
}

func (s *AuthSuite) TestLogin() {
	s.MakeRequest("POST", "/api/v1/auth/register", map[string]string{
		"email":    "login@test.com",
		"password": "Password123!",
	}, "")

	res := s.MakeRequest("POST", "/api/v1/auth/login", map[string]string{
		"email":    "login@test.com",
		"password": "Password123!",
	}, "")
	s.Equal(http.StatusOK, res.Code)
	s.Contains(res.Body.String(), "token")

	resFail := s.MakeRequest("POST", "/api/v1/auth/login", map[string]string{
		"email":    "login@test.com",
		"password": "WrongPassword",
	}, "")
	s.Equal(http.StatusUnauthorized, resFail.Code)
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthSuite))
}
