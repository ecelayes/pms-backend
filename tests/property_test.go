package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type PropertySuite struct {
	BaseSuite
	token string
	orgID string
}

func (s *PropertySuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.token, s.orgID = s.GetAdminTokenAndOrg()
}

func (s *PropertySuite) TestCRUDProperty() {
	res := s.MakeRequest("POST", "/api/v1/properties", map[string]string{
		"organization_id": s.orgID,
		"name":            "Property Test",
		"code":            "PRP",
		"type":            "HOTEL",
	}, s.token)
	s.Require().Equal(http.StatusCreated, res.Code)

	var data map[string]string
	json.Unmarshal(res.Body.Bytes(), &data)
	propertyID := data["property_id"]

	resList := s.MakeRequest("GET", "/api/v1/properties?organization_id="+s.orgID, nil, s.token)
	s.Equal(http.StatusOK, resList.Code)
	
	var response entity.PaginatedResponse[entity.Property]
	json.Unmarshal(resList.Body.Bytes(), &response)
	s.NotEmpty(response.Data)
	found := false
	for _, p := range response.Data {
		if p.ID == propertyID {
			found = true
			break
		}
	}
	s.True(found, "Newly created property should be in the list")

	resDel := s.MakeRequest("DELETE", "/api/v1/properties/"+propertyID, nil, s.token)
	s.Equal(http.StatusOK, resDel.Code)
}

func (s *PropertySuite) TestPropertyValidation() {
	res := s.MakeRequest("POST", "/api/v1/properties", map[string]string{
		"organization_id": s.orgID,
		"name":            "",
		"code":            "INV",
	}, s.token)
	s.Equal(http.StatusBadRequest, res.Code)

	res2 := s.MakeRequest("POST", "/api/v1/properties", map[string]string{
		"organization_id": s.orgID,
		"name":            "Valid Name",
		"code":            "",
	}, s.token)
	s.Equal(http.StatusBadRequest, res2.Code)
}

func (s *PropertySuite) TestPropertyNotFound() {
	res := s.MakeRequest("GET", "/api/v1/properties/00000000-0000-0000-0000-000000000000", nil, s.token)
	s.Equal(http.StatusNotFound, res.Code)

	resUpd := s.MakeRequest("PUT", "/api/v1/properties/00000000-0000-0000-0000-000000000000", map[string]string{
		"name": "Updated Name",
	}, s.token)
	s.Equal(http.StatusNotFound, resUpd.Code)

	resDel := s.MakeRequest("DELETE", "/api/v1/properties/00000000-0000-0000-0000-000000000000", nil, s.token)
	s.Equal(http.StatusNotFound, resDel.Code)
}

func (s *PropertySuite) TestPropertyDuplicateCode() {
	res := s.MakeRequest("POST", "/api/v1/properties", map[string]string{
		"organization_id": s.orgID,
		"name":            "Unique Property",
		"code":            "UNI",
	}, s.token)
	s.Equal(http.StatusCreated, res.Code)

	res2 := s.MakeRequest("POST", "/api/v1/properties", map[string]string{
		"organization_id": s.orgID,
		"name":            "Another Property",
		"code":            "UNI",
	}, s.token)
	s.Equal(http.StatusConflict, res2.Code)
}

func TestPropertySuite(t *testing.T) {
	suite.Run(t, new(PropertySuite))
}
