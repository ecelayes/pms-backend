package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UserSuite struct {
	BaseSuite
	ownerToken string
	orgID      string
}

func (s *UserSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.ownerToken, s.orgID = s.GetAdminTokenAndOrg()
}

func (s *UserSuite) TestUserHierarchy() {
	res := s.MakeRequest("POST", "/api/v1/users", map[string]string{
		"organization_id": s.orgID,
		"email":           "manager@corp.com",
		"password":        "secret123",
		"role":            "manager",
		"first_name":      "Manager",
    "last_name":       "One",
    "phone":           "987654",
	}, s.ownerToken)
	s.Equal(http.StatusCreated, res.Code)

	resFail := s.MakeRequest("POST", "/api/v1/users", map[string]string{
		"organization_id": s.orgID,
		"email":           "another_owner@corp.com",
		"password":        "secret123",
		"role":            "owner",
		"first_name":      "Owner",
    "last_name":       "Two",
    "phone":           "987654",
	}, s.ownerToken)
	
	s.Equal(http.StatusForbidden, resFail.Code)
	s.Contains(resFail.Body.String(), "only super_admin can create organization owners")
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserSuite))
}
