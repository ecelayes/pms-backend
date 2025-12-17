package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/ecelayes/pms-backend/internal/entity"
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
	createData := map[string]interface{}{
		"name": "Test Org",
	}
	res := s.MakeRequest("POST", "/api/v1/organizations", createData, s.superToken)
	s.Equal(http.StatusCreated, res.Code)

	var createResp map[string]string
	json.Unmarshal(res.Body.Bytes(), &createResp)
	orgID := createResp["organization_id"]
	s.NotEmpty(orgID)

	res = s.MakeRequest("GET", "/api/v1/organizations/"+orgID, nil, s.superToken)
	s.Equal(http.StatusOK, res.Code)

	var org entity.Organization
	json.Unmarshal(res.Body.Bytes(), &org)
	s.Equal("Test Org", org.Name)
	s.NotEmpty(org.Code)

	res = s.MakeRequest("GET", "/api/v1/organizations?page=1&limit=10", nil, s.superToken)
	s.Equal(http.StatusOK, res.Code)

	var listResp entity.PaginatedResponse[entity.Organization]
	err := json.Unmarshal(res.Body.Bytes(), &listResp)
	s.NoError(err)

	found := false
	for _, o := range listResp.Data {
		if o.ID == orgID {
			found = true
			break
		}
	}
	s.True(found, "Should find created organization in list")
	s.GreaterOrEqual(listResp.Meta.TotalItems, int64(1))

	resUnlim := s.MakeRequest("GET", "/api/v1/organizations", nil, s.superToken)
	s.Equal(http.StatusOK, resUnlim.Code)
	
	var listUnlim entity.PaginatedResponse[entity.Organization]
	json.Unmarshal(resUnlim.Body.Bytes(), &listUnlim)
	
	s.Equal(1, listUnlim.Meta.Page)
	s.Equal(1, listUnlim.Meta.TotalPages)
	s.Equal(listUnlim.Meta.TotalItems, int64(len(listUnlim.Data)))
	s.GreaterOrEqual(len(listUnlim.Data), 1)

	updateData := map[string]interface{}{
		"name": "Updated Org",
	}
	res = s.MakeRequest("PUT", "/api/v1/organizations/"+orgID, updateData, s.superToken)
	s.Equal(http.StatusOK, res.Code)

	res = s.MakeRequest("GET", "/api/v1/organizations/"+orgID, nil, s.superToken)
	json.Unmarshal(res.Body.Bytes(), &org)
	s.Equal("Updated Org", org.Name)

	res = s.MakeRequest("DELETE", "/api/v1/organizations/"+orgID, nil, s.superToken)
	s.Equal(http.StatusOK, res.Code)

	res = s.MakeRequest("GET", "/api/v1/organizations/"+orgID, nil, s.superToken)
	s.Equal(http.StatusNotFound, res.Code)
}

func (s *OrganizationSuite) TestOrganizationValidation() {
	res := s.MakeRequest("POST", "/api/v1/organizations", map[string]interface{}{
		"name": "",
	}, s.superToken)
	s.Equal(http.StatusBadRequest, res.Code)
}

func (s *OrganizationSuite) TestResourceNotFound() {
	res := s.MakeRequest("GET", "/api/v1/organizations/invalid-uuid", nil, s.superToken)
	s.Equal(http.StatusNotFound, res.Code)

	res2 := s.MakeRequest("GET", "/api/v1/organizations/00000000-0000-0000-0000-000000000000", nil, s.superToken)
	s.Equal(http.StatusNotFound, res2.Code)
}

func TestOrganizationSuite(t *testing.T) {
	suite.Run(t, new(OrganizationSuite))
}
