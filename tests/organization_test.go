package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type OrganizationSuite struct {
	BaseSuite
	token string
}

func (s *OrganizationSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.token, _ = s.GetAdminTokenAndOrg()
}

func (s *OrganizationSuite) TestCRUDOrganization() {
	res := s.MakeRequest("POST", "/api/v1/organizations", map[string]string{
		"name": "Hilton Group",
	}, s.token)
	s.Equal(http.StatusCreated, res.Code)

	var data map[string]string
	json.Unmarshal(res.Body.Bytes(), &data)
	orgID := data["organization_id"]
	s.NotEmpty(orgID)

	resGet := s.MakeRequest("GET", "/api/v1/organizations/"+orgID, nil, s.token) // <--- TOKEN
	s.Equal(http.StatusOK, resGet.Code)
	s.Contains(resGet.Body.String(), "Hilton Group")

	resUpdate := s.MakeRequest("PUT", "/api/v1/organizations/"+orgID, map[string]string{
		"name": "Hilton International",
	}, s.token)
	s.Equal(http.StatusOK, resUpdate.Code)

	resDelete := s.MakeRequest("DELETE", "/api/v1/organizations/"+orgID, nil, s.token) // <--- TOKEN
	s.Equal(http.StatusOK, resDelete.Code)
}

func TestOrganizationSuite(t *testing.T) {
	suite.Run(t, new(OrganizationSuite))
}
