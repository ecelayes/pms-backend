package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RoomSuite struct {
	BaseSuite
	token   string
	orgID   string
	hotelID string
}

func (s *RoomSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.token, s.orgID = s.GetAdminTokenAndOrg()

	res := s.MakeRequest("POST", "/api/v1/hotels", map[string]string{
		"organization_id": s.orgID,
		"name":            "Room Hotel",
		"code":            "RHO",
	}, s.token)
	s.Require().Equal(http.StatusCreated, res.Code)
	
	var data map[string]string
	json.Unmarshal(res.Body.Bytes(), &data)
	s.hotelID = data["hotel_id"]
}

func (s *RoomSuite) TestCRUDRoom() {
	res := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID,
		"name":           "Suite",
		"code":           "SUI",
		"total_quantity": 5,
		"base_price":     150.0,
		"max_occupancy":  4,
		"max_adults":     2,
		"max_children":   2,
		"amenities":      []string{"wifi"},
	}, s.token)
	s.Equal(http.StatusCreated, res.Code)
	
	var data map[string]string
	json.Unmarshal(res.Body.Bytes(), &data)
	id := data["room_type_id"]

	resGet := s.MakeRequest("GET", "/api/v1/room-types/"+id, nil, s.token)
	s.Equal(http.StatusOK, resGet.Code)
	
	var roomMap map[string]interface{}
	json.Unmarshal(resGet.Body.Bytes(), &roomMap)
	s.Equal(150.0, roomMap["base_price"])

	resUpd := s.MakeRequest("PUT", "/api/v1/room-types/"+id, map[string]interface{}{
		"base_price": 200.0,
	}, s.token)
	s.Equal(http.StatusOK, resUpd.Code)

	resGet2 := s.MakeRequest("GET", "/api/v1/room-types/"+id, nil, s.token)
	json.Unmarshal(resGet2.Body.Bytes(), &roomMap)
	s.Equal(200.0, roomMap["base_price"])
}

func TestRoomSuite(t *testing.T) {
	suite.Run(t, new(RoomSuite))
}
