package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ReservationSuite struct {
	BaseSuite
	token      string
	roomTypeID string
}

func (s *ReservationSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.token = s.GetAdminToken()

	resH := s.MakeRequest("POST", "/api/v1/hotels", map[string]string{"name": "H", "code": "HHH"}, s.token)
	var dataH map[string]string
	json.Unmarshal(resH.Body.Bytes(), &dataH)
	
	resR := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id": dataH["hotel_id"], "name": "R", "code": "RRR", "total_quantity": 2,
	}, s.token)
	var dataR map[string]string
	json.Unmarshal(resR.Body.Bytes(), &dataR)
	s.roomTypeID = dataR["room_type_id"]

	s.MakeRequest("POST", "/api/v1/pricing/rules", map[string]interface{}{
		"room_type_id": s.roomTypeID,
		"start": "2025-01-01", "end": "2025-01-10", "price": 100.0, "priority": 0,
	}, s.token)
}

func (s *ReservationSuite) TestReservationLifecycle() {
	res := s.MakeRequest("POST", "/api/v1/reservations", map[string]interface{}{
		"room_type_id": s.roomTypeID,
		"guest_email":  "guest@test.com",
		"start":        "2025-01-01",
		"end":          "2025-01-05",
	}, "")
	s.Equal(http.StatusCreated, res.Code)

	var data map[string]string
	json.Unmarshal(res.Body.Bytes(), &data)
	resCode := data["reservation_code"]

	resGet := s.MakeRequest("GET", "/api/v1/reservations/"+resCode, nil, "")
	s.Equal(http.StatusOK, resGet.Code)
	
	var resData map[string]interface{}
	json.Unmarshal(resGet.Body.Bytes(), &resData)
	id := resData["id"].(string)

	resDelete := s.MakeRequest("DELETE", "/api/v1/reservations/"+id, nil, s.token)
	s.Equal(http.StatusOK, resDelete.Code)

	resGet2 := s.MakeRequest("GET", "/api/v1/reservations/"+resCode, nil, "")
	s.Equal(http.StatusNotFound, resGet2.Code)
}

func TestReservationSuite(t *testing.T) {
	suite.Run(t, new(ReservationSuite))
}
