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
		"name":           "Suite",
		"code":           "SUI",
		"total_quantity": 5,
	}, s.token)
	s.Equal(http.StatusCreated, res.Code)
	
	var data map[string]string
	json.Unmarshal(res.Body.Bytes(), &data)
	roomID := data["room_type_id"]

	resUpdate := s.MakeRequest("PUT", "/api/v1/room-types/"+roomID, map[string]interface{}{
		"name":           "Suite Updated",
		"code":           "UPX",
		"total_quantity": 10,
	}, s.token)
	s.Equal(http.StatusOK, resUpdate.Code)

	resDelete := s.MakeRequest("DELETE", "/api/v1/room-types/"+roomID, nil, s.token)
	s.Equal(http.StatusOK, resDelete.Code)

	resPrice := s.MakeRequest("POST", "/api/v1/pricing/rules", map[string]interface{}{
		"room_type_id": roomID,
		"start": "2025-01-01", "end": "2025-01-02", "price": 100.0, "priority": 0,
	}, s.token)
	s.Equal(http.StatusNotFound, resPrice.Code)
}

func TestRoomSuite(t *testing.T) {
	suite.Run(t, new(RoomSuite))
}
