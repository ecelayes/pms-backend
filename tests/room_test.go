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
	hotelID string
}

func (s *RoomSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.token = s.GetAdminToken()

	res := s.MakeRequest("POST", "/api/v1/hotels", map[string]string{
		"name": "Room Test Hotel", "code": "RTH",
	}, s.token)
	var data map[string]string
	json.Unmarshal(res.Body.Bytes(), &data)
	s.hotelID = data["hotel_id"]
}

func (s *RoomSuite) TestCRUDRoom() {
	res := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID,
		"name":           "Family Suite",
		"code":           "FAM",
		"total_quantity": 5,
		"max_occupancy":  4,
		"max_adults":     2,
		"max_children":   2,
		"amenities":      []string{"wifi", "tv", "kitchen"},
	}, s.token)
	s.Equal(http.StatusCreated, res.Code)
	
	var data map[string]string
	json.Unmarshal(res.Body.Bytes(), &data)
	roomID := data["room_type_id"]

	resUpdate := s.MakeRequest("PUT", "/api/v1/room-types/"+roomID, map[string]interface{}{
		"name":           "Super Family Suite",
		"code":           "SFM",
		"total_quantity": 10,
		"max_occupancy":  5,
		"max_adults":     3,
		"max_children":   2,
		"amenities":      []string{"wifi", "tv", "jacuzzi"},
	}, s.token)
	s.Equal(http.StatusOK, resUpdate.Code)

	resDelete := s.MakeRequest("DELETE", "/api/v1/room-types/"+roomID, nil, s.token)
	s.Equal(http.StatusOK, resDelete.Code)
}

func TestRoomSuite(t *testing.T) {
	suite.Run(t, new(RoomSuite))
}
