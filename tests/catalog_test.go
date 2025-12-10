package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CatalogSuite struct {
	BaseSuite
	superToken string
	ownerToken string
}

func (s *CatalogSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.superToken = s.GetSuperAdminToken()
	s.ownerToken, _ = s.GetAdminTokenAndOrg()
}

func (s *CatalogSuite) TestAmenitiesLifecycle() {
	resCreate := s.MakeRequest("POST", "/api/v1/amenities", map[string]string{
		"name": "WiFi High Speed",
		"icon": "wifi",
	}, s.superToken)
	s.Equal(http.StatusCreated, resCreate.Code)
	
	var data map[string]string
	json.Unmarshal(resCreate.Body.Bytes(), &data)
	id := data["id"]

	resForbidden := s.MakeRequest("POST", "/api/v1/amenities", map[string]string{
		"name": "Pool",
	}, s.ownerToken)
	s.Equal(http.StatusForbidden, resForbidden.Code)

	resList := s.MakeRequest("GET", "/api/v1/amenities", nil, s.ownerToken)
	s.Equal(http.StatusOK, resList.Code)
	s.Contains(resList.Body.String(), "WiFi High Speed")

	resUpdate := s.MakeRequest("PUT", "/api/v1/amenities/"+id, map[string]string{
		"name": "WiFi 6E",
	}, s.superToken)
	s.Equal(http.StatusOK, resUpdate.Code)

	resDel := s.MakeRequest("DELETE", "/api/v1/amenities/"+id, nil, s.superToken)
	s.Equal(http.StatusOK, resDel.Code)
}

func (s *CatalogSuite) TestServicesLifecycle() {
	res := s.MakeRequest("POST", "/api/v1/services", map[string]string{"name": "Parking"}, s.superToken)
	s.Equal(http.StatusCreated, res.Code)
	
	resGet := s.MakeRequest("GET", "/api/v1/services", nil, s.ownerToken)
	s.Equal(http.StatusOK, resGet.Code)
	s.Contains(resGet.Body.String(), "Parking")
}

func TestCatalogSuite(t *testing.T) {
	suite.Run(t, new(CatalogSuite))
}
