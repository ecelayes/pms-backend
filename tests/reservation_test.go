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
	orgID      string
	roomTypeID string
}

func (s *ReservationSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.token, s.orgID = s.GetAdminTokenAndOrg()

	resH := s.MakeRequest("POST", "/api/v1/hotels", map[string]string{
		"organization_id": s.orgID,
		"name":            "Res Hotel",
		"code":            "RHO",
	}, s.token)
	s.Require().Equal(http.StatusCreated, resH.Code)

	var dataH map[string]string
	json.Unmarshal(resH.Body.Bytes(), &dataH)
	
	resR := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       dataH["hotel_id"], 
		"name":           "Std", "code": "STD", 
		"total_quantity": 5,
		"max_occupancy":  4, "max_adults": 2, "max_children": 2,
		"amenities":      []string{"wifi"},
	}, s.token)
	s.Require().Equal(http.StatusCreated, resR.Code)

	var dataR map[string]string
	json.Unmarshal(resR.Body.Bytes(), &dataR)
	s.roomTypeID = dataR["room_type_id"]

	resP := s.MakeRequest("POST", "/api/v1/pricing/rules", map[string]interface{}{
		"room_type_id": s.roomTypeID,
		"start": "2025-01-01", "end": "2025-01-10", "price": 100.0, "priority": 0,
	}, s.token)
	s.Require().Equal(http.StatusCreated, resP.Code)
}

func (s *ReservationSuite) TestReservationCRUD() {
	res := s.MakeRequest("POST", "/api/v1/reservations", map[string]interface{}{
		"room_type_id":     s.roomTypeID,
		"guest_email":      "guest@test.com",
		"guest_first_name": "John", "guest_last_name": "Doe",
		"start":            "2025-01-01", "end": "2025-01-05",
		"adults":           2, "children": 0,
	}, "")
	
	s.Require().Equal(http.StatusCreated, res.Code, "Response: "+res.Body.String())

	var data map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &data)
	code := data["reservation_code"].(string)

	resGet := s.MakeRequest("GET", "/api/v1/reservations/"+code, nil, "")
	s.Equal(http.StatusOK, resGet.Code)
	
	bodyString := resGet.Body.String()
	s.Contains(bodyString, `"guest_id"`)
	s.Contains(bodyString, `"status":"confirmed"`)
	s.Contains(bodyString, code)
}

func TestReservationSuite(t *testing.T) {
	suite.Run(t, new(ReservationSuite))
}
