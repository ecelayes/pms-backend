package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type CatalogSuite struct {
	BaseSuite
	superToken string
}

func (s *CatalogSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.superToken = s.GetSuperAdminToken()
}

func (s *CatalogSuite) TestCRUDAmenity() {
	createData := map[string]interface{}{
		"name":        "Wifi",
		"description": "High speed internet",
		"icon":        "wifi-icon",
	}
	res := s.MakeRequest("POST", "/api/v1/amenities", createData, s.superToken)
	s.Equal(http.StatusCreated, res.Code)

	var createResp map[string]string
	json.Unmarshal(res.Body.Bytes(), &createResp)
	amenityID := createResp["id"]
	s.NotEmpty(amenityID)

	res = s.MakeRequest("GET", "/api/v1/amenities/"+amenityID, nil, s.superToken)
	s.Equal(http.StatusOK, res.Code)
	
	var a entity.Amenity
	json.Unmarshal(res.Body.Bytes(), &a)
	s.Equal("Wifi", a.Name)

	res = s.MakeRequest("GET", "/api/v1/amenities?page=1&limit=5", nil, s.superToken)
	s.Equal(http.StatusOK, res.Code)

	var listResp entity.PaginatedResponse[entity.Amenity]
	err := json.Unmarshal(res.Body.Bytes(), &listResp)
	s.NoError(err)
	
	found := false
	for _, item := range listResp.Data {
		if item.ID == amenityID {
			found = true
			break
		}
	}
	s.True(found, "Should find created amenity in list")
	s.NotEmpty(listResp.Meta.TotalItems)

	resUnlim := s.MakeRequest("GET", "/api/v1/amenities", nil, s.superToken)
	s.Equal(http.StatusOK, resUnlim.Code)
	var listUnlim entity.PaginatedResponse[entity.Amenity]
	json.Unmarshal(resUnlim.Body.Bytes(), &listUnlim)
	s.Equal(1, listUnlim.Meta.TotalPages)
	s.Equal(listUnlim.Meta.TotalItems, int64(len(listUnlim.Data)))

	updateData := map[string]interface{}{
		"name":        "Free Wifi",
		"description": "Free high speed internet",
		"icon":        "wifi-free",
	}
	res = s.MakeRequest("PUT", "/api/v1/amenities/"+amenityID, updateData, s.superToken)
	s.Equal(http.StatusOK, res.Code)

	res = s.MakeRequest("DELETE", "/api/v1/amenities/"+amenityID, nil, s.superToken)
	s.Equal(http.StatusOK, res.Code)
}

func (s *CatalogSuite) TestCRUDService() {
	createData := map[string]interface{}{
		"name":        "Spa",
		"description": "Relaxing spa",
		"icon":        "spa-icon",
	}
	res := s.MakeRequest("POST", "/api/v1/services", createData, s.superToken)
	s.Equal(http.StatusCreated, res.Code)

	var createResp map[string]string
	json.Unmarshal(res.Body.Bytes(), &createResp)
	serviceID := createResp["id"]
	s.NotEmpty(serviceID)

	res = s.MakeRequest("GET", "/api/v1/services/"+serviceID, nil, s.superToken)
	s.Equal(http.StatusOK, res.Code)
	
	var serv entity.HotelService
	json.Unmarshal(res.Body.Bytes(), &serv)
	s.Equal("Spa", serv.Name)

	res = s.MakeRequest("GET", "/api/v1/services?page=1&limit=5", nil, s.superToken)
	s.Equal(http.StatusOK, res.Code)

	var listResp entity.PaginatedResponse[entity.HotelService]
	err := json.Unmarshal(res.Body.Bytes(), &listResp)
	s.NoError(err)
	
	found := false
	for _, item := range listResp.Data {
		if item.ID == serviceID {
			found = true
			break
		}
	}
	s.True(found, "Should find created service in list")

	resUnlim := s.MakeRequest("GET", "/api/v1/services", nil, s.superToken)
	s.Equal(http.StatusOK, resUnlim.Code)
	var listUnlim entity.PaginatedResponse[entity.HotelService]
	json.Unmarshal(resUnlim.Body.Bytes(), &listUnlim)
	s.Equal(1, listUnlim.Meta.TotalPages)
	s.Equal(listUnlim.Meta.TotalItems, int64(len(listUnlim.Data)))

	updateData := map[string]interface{}{
		"name":        "Luxury Spa",
		"description": "Very relaxing spa",
		"icon":        "spa-lux",
	}
	res = s.MakeRequest("PUT", "/api/v1/services/"+serviceID, updateData, s.superToken)
	s.Equal(http.StatusOK, res.Code)

	res = s.MakeRequest("DELETE", "/api/v1/services/"+serviceID, nil, s.superToken)
	s.Equal(http.StatusOK, res.Code)
}

func (s *CatalogSuite) TestCatalogValidation() {
	res := s.MakeRequest("POST", "/api/v1/amenities", map[string]interface{}{
		"name": "", "icon": "wifi",
	}, s.superToken)
	s.Equal(http.StatusBadRequest, res.Code)

	res2 := s.MakeRequest("POST", "/api/v1/services", map[string]interface{}{
		"name": "", "icon": "spa",
	}, s.superToken)
	s.Equal(http.StatusBadRequest, res2.Code)
}

func (s *CatalogSuite) TestResourceNotFound() {
	res := s.MakeRequest("GET", "/api/v1/amenities/00000000-0000-0000-0000-000000000000", nil, s.superToken)
	s.Equal(http.StatusNotFound, res.Code)

	res2 := s.MakeRequest("GET", "/api/v1/services/00000000-0000-0000-0000-000000000000", nil, s.superToken)
	s.Equal(http.StatusNotFound, res2.Code)
}

func (s *CatalogSuite) TestCatalogDuplicate() {
	res := s.MakeRequest("POST", "/api/v1/amenities", map[string]interface{}{
		"name": "Pool", "icon": "pool",
	}, s.superToken)
	s.Equal(http.StatusCreated, res.Code)

	res2 := s.MakeRequest("POST", "/api/v1/amenities", map[string]interface{}{
		"name": "Pool", "icon": "pool2",
	}, s.superToken)
	s.Equal(http.StatusConflict, res2.Code)

	res3 := s.MakeRequest("POST", "/api/v1/services", map[string]interface{}{
		"name": "Massage", "icon": "hand",
	}, s.superToken)
	s.Equal(http.StatusCreated, res3.Code)

	res4 := s.MakeRequest("POST", "/api/v1/services", map[string]interface{}{
		"name": "Massage", "icon": "hand2",
	}, s.superToken)
	s.Equal(http.StatusConflict, res4.Code)
}

func TestCatalogSuite(t *testing.T) {
	suite.Run(t, new(CatalogSuite))
}
