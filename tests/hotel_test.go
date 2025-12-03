package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type HotelSuite struct {
	BaseSuite
	token string
}

func (s *HotelSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.token = s.GetAdminToken()
}

func (s *HotelSuite) TestCRUDHotel() {
	resCreate := s.MakeRequest("POST", "/api/v1/hotels", map[string]string{
		"name": "Original Name",
		"code": "ORG",
	}, s.token)
	s.Equal(http.StatusCreated, resCreate.Code)
	
	var data map[string]string
	json.Unmarshal(resCreate.Body.Bytes(), &data)
	hotelID := data["hotel_id"]

	resUpdate := s.MakeRequest("PUT", "/api/v1/hotels/"+hotelID, map[string]string{
		"name": "Updated Name",
		"code": "UPD",
	}, s.token)
	s.Equal(http.StatusOK, resUpdate.Code)

	resList := s.MakeRequest("GET", "/api/v1/hotels", nil, s.token)
	s.Contains(resList.Body.String(), "Updated Name")
	s.Contains(resList.Body.String(), "UPD")

	resDelete := s.MakeRequest("DELETE", "/api/v1/hotels/"+hotelID, nil, s.token)
	s.Equal(http.StatusOK, resDelete.Code)

	resListAfter := s.MakeRequest("GET", "/api/v1/hotels", nil, s.token)
	s.NotContains(resListAfter.Body.String(), hotelID, "Deleted hotel should not be listed")
	
	resDelete2 := s.MakeRequest("DELETE", "/api/v1/hotels/"+hotelID, nil, s.token)
	s.NotEqual(http.StatusOK, resDelete2.Code)
}

func TestHotelSuite(t *testing.T) {
	suite.Run(t, new(HotelSuite))
}
