package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type ReservationSuite struct {
	BaseSuite
	token      string
	orgID      string
	hotelID    string
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
	s.hotelID = dataH["hotel_id"]

	resR := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID, 
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

func (s *ReservationSuite) TestReservationWithMealPlan() {
	resRP := s.MakeRequest("POST", "/api/v1/rate-plans", map[string]interface{}{
		"hotel_id":     s.hotelID,
		"room_type_id": s.roomTypeID,
		"name":         "Plus Breakfast",
		"meal_plan": map[string]interface{}{
			"included":      true,
			"price_per_pax": 20.0,
		},
		"cancellation_policy": map[string]interface{}{"is_refundable": true},
		"payment_policy":      map[string]interface{}{"timing": 0},
	}, s.token)
	s.Require().Equal(http.StatusCreated, resRP.Code)
	
	var dataRP map[string]string
	json.Unmarshal(resRP.Body.Bytes(), &dataRP)
	planID := dataRP["rate_plan_id"]

	res := s.MakeRequest("POST", "/api/v1/reservations", map[string]interface{}{
		"room_type_id":     s.roomTypeID,
		"rate_plan_id":     planID,
		"guest_email":      "meal@test.com",
		"guest_first_name": "Meal", "guest_last_name": "Tester",
		"start":            "2025-01-01", "end": "2025-01-04",
		"adults":           2, "children": 0,
	}, "")
	
	s.Require().Equal(http.StatusCreated, res.Code)

	var dataRes map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &dataRes)
	code := dataRes["reservation_code"].(string)

	resGet := s.MakeRequest("GET", "/api/v1/reservations/"+code, nil, "")
	s.Equal(http.StatusOK, resGet.Code)
	
	var resData entity.Reservation
	json.Unmarshal(resGet.Body.Bytes(), &resData)
	
	s.Equal(420.0, resData.TotalPrice, "El precio total debe incluir el recargo de desayuno")
	s.NotNil(resData.RatePlanID)
	s.Equal(planID, *resData.RatePlanID)
}

func TestReservationSuite(t *testing.T) {
	suite.Run(t, new(ReservationSuite))
}
