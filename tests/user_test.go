package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UserSuite struct {
	BaseSuite
	adminToken string
	orgID      string
}

func (s *UserSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.adminToken, s.orgID = s.GetAdminTokenAndOrg()
}

func (s *UserSuite) TestCreateUserLinkedToOrg() {
	res := s.MakeRequest("POST", "/api/v1/users", map[string]string{
		"organization_id": s.orgID,
		"email":           "manager@corp.com",
		"password":        "secret123",
		"role":            "manager",
	}, s.adminToken)
	
	s.Require().Equal(http.StatusCreated, res.Code)

	var data map[string]string
	json.Unmarshal(res.Body.Bytes(), &data)
	userID := data["user_id"]
	s.NotEmpty(userID)

	resLogin := s.MakeRequest("POST", "/api/v1/auth/login", map[string]string{
		"email":    "manager@corp.com",
		"password": "secret123",
	}, "")
	
	s.Equal(http.StatusOK, resLogin.Code)
	s.Contains(resLogin.Body.String(), "token")
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserSuite))
}
