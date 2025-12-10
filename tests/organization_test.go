package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type OrganizationSuite struct {
	BaseSuite
	superToken string
}

func (s *OrganizationSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.superToken = s.GetSuperAdminToken()
}

func (s *OrganizationSuite) TestCRUDOrganization() {
	res := s.MakeRequest("POST", "/api/v1/organizations", map[string]string{
		"name": "Hilton Group",
	}, s.superToken)
	s.Equal(http.StatusCreated, res.Code)

	var data map[string]string
	json.Unmarshal(res.Body.Bytes(), &data)
	orgID := data["organization_id"]

	resGet := s.MakeRequest("GET", "/api/v1/organizations/"+orgID, nil, s.superToken)
	s.Equal(http.StatusOK, resGet.Code)
	s.Contains(resGet.Body.String(), "Hilton Group")

	resDel := s.MakeRequest("DELETE", "/api/v1/organizations/"+orgID, nil, s.superToken)
	s.Equal(http.StatusOK, resDel.Code)
}

func TestOrganizationSuite(t *testing.T) {
	suite.Run(t, new(OrganizationSuite))
}
