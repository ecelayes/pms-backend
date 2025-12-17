package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/ecelayes/pms-backend/internal/entity"
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

func (s *RoomSuite) TestListRooms() {
	s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID,
		"name":           "List Test Room",
		"code":           "LTR",
		"total_quantity": 10,
		"base_price":     100.0,
		"max_occupancy":  2, "max_adults": 2, "max_children": 0,
		"amenities":      []string{"wifi"},
	}, s.token)

	res := s.MakeRequest("GET", "/api/v1/room-types?hotel_id="+s.hotelID+"&page=1&limit=5", nil, s.token)
	s.Equal(http.StatusOK, res.Code)

	var response entity.PaginatedResponse[entity.RoomType]
	json.Unmarshal(res.Body.Bytes(), &response)

	s.NotEmpty(response.Data)
	s.Equal(1, response.Meta.Page)
	s.Equal(5, response.Meta.Limit)
	s.GreaterOrEqual(response.Meta.TotalItems, int64(1))
}

func (s *RoomSuite) TestRoomTypeValidation() {
	res := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID,
		"name":           "Negative Room",
		"code":           "NEG",
		"base_price":     -10.0,
		"max_occupancy":  2,
	}, s.token)
	s.Equal(http.StatusBadRequest, res.Code)

	res2 := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID,
		"name":           "Zero Cap Room",
		"code":           "ZCP",
		"base_price":     100.0,
		"max_occupancy":  0,
	}, s.token)
	s.Equal(http.StatusBadRequest, res2.Code)

	res3 := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID,
		"name":           "",
		"code":           "EMP",
		"base_price":     100.0,
		"max_occupancy":  2,
	}, s.token)
	s.Equal(http.StatusBadRequest, res3.Code)
}

func (s *RoomSuite) TestRoomTypeNotFound() {
	res := s.MakeRequest("GET", "/api/v1/room-types/00000000-0000-0000-0000-000000000000", nil, s.token)
	s.Equal(http.StatusNotFound, res.Code)

	resUpd := s.MakeRequest("PUT", "/api/v1/room-types/00000000-0000-0000-0000-000000000000", map[string]interface{}{
		"base_price": 200.0,
	}, s.token)
	s.Equal(http.StatusNotFound, resUpd.Code)
}

func (s *RoomSuite) TestRoomTypeDuplicateCode() {
	res := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID,
		"name":           "Unique Room",
		"code":           "UNI",
		"base_price":     100.0,
		"total_quantity": 5,
		"max_occupancy":  2,
		"max_adults":     2,
		"max_children":   0,
		"amenities":      []string{"wifi"},
	}, s.token)
	s.Equal(http.StatusCreated, res.Code)

	res2 := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID,
		"name":           "Another Room",
		"code":           "UNI",
		"base_price":     120.0,
		"total_quantity": 5,
		"max_occupancy":  2,
		"max_adults":     2,
		"max_children":   0,
		"amenities":      []string{"wifi"},
	}, s.token)
	s.Equal(http.StatusConflict, res2.Code)
}

func TestRoomSuite(t *testing.T) {
	suite.Run(t, new(RoomSuite))
}
