package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LifecycleSuite struct {
	BaseSuite
	ownerToken string
	hotelID    string
	roomID     string
	resCode    string
}

func (s *LifecycleSuite) TestFullLifecycle() {
	s.Run("1. Register Owner", func() {
		res := s.MakeRequest("POST", "/api/v1/auth/register", map[string]string{
			"email":    "ceo@chain.com",
			"password": "SecurePass123!",
		}, "")
		s.Equal(http.StatusCreated, res.Code)
	})

	s.Run("2. Login Owner", func() {
		res := s.MakeRequest("POST", "/api/v1/auth/login", map[string]string{
			"email":    "ceo@chain.com",
			"password": "SecurePass123!",
		}, "")
		s.Equal(http.StatusOK, res.Code)

		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.ownerToken = data["token"]
		s.NotEmpty(s.ownerToken)
	})

	s.Run("3. Create Hotel", func() {
		res := s.MakeRequest("POST", "/api/v1/hotels", map[string]string{
			"name": "Grand Lifecycle Hotel",
			"code": "LIF",
		}, s.ownerToken)
		s.Equal(http.StatusCreated, res.Code)

		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.hotelID = data["hotel_id"]
		s.NotEmpty(s.hotelID)
	})

	s.Run("4. Create Room", func() {
		res := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
			"hotel_id":       s.hotelID,
			"name":           "Suite Lifecycle",
			"code":           "SUI",
			"total_quantity": 5,
		}, s.ownerToken)
		s.Equal(http.StatusCreated, res.Code)

		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.roomID = data["room_type_id"]
		s.NotEmpty(s.roomID)
	})

	s.Run("5. Set Pricing", func() {
		res := s.MakeRequest("POST", "/api/v1/pricing/rules", map[string]interface{}{
			"room_type_id": s.roomID,
			"start":        "2025-10-01",
			"end":          "2025-10-05",
			"price":        200.0,
			"priority":     10,
		}, s.ownerToken)
		s.Equal(http.StatusCreated, res.Code)
	})

	s.Run("6. Customer Reserves", func() {
		res := s.MakeRequest("POST", "/api/v1/reservations", map[string]interface{}{
			"room_type_id": s.roomID,
			"guest_email":  "tourist@gmail.com",
			"start":        "2025-10-01",
			"end":          "2025-10-05",
		}, "")
		s.Equal(http.StatusCreated, res.Code)

		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.resCode = data["reservation_code"]
		s.Contains(s.resCode, "LIF-SUI-")
	})

	s.Run("7. Check Availability", func() {
		res := s.MakeRequest("GET", "/api/v1/availability?start=2025-10-01&end=2025-10-05", nil, "")
		s.Equal(http.StatusOK, res.Code)
		s.Contains(res.Body.String(), `"available_qty":4`)
	})

	s.Run("8. Cancel Reservation", func() {
		resGet := s.MakeRequest("GET", "/api/v1/reservations/"+s.resCode, nil, "")
		var dataGet map[string]interface{}
		json.Unmarshal(resGet.Body.Bytes(), &dataGet)
		id := dataGet["id"].(string)

		resCancel := s.MakeRequest("POST", "/api/v1/reservations/"+id+"/cancel", nil, "")
		s.Equal(http.StatusOK, resCancel.Code)
	})

	s.Run("9. Check Restored Availability", func() {
		res := s.MakeRequest("GET", "/api/v1/availability?start=2025-10-01&end=2025-10-05", nil, "")
		s.Contains(res.Body.String(), `"available_qty":5`) // Volvió a 5
	})
	
	s.Run("10. Delete Hotel", func() {
		res := s.MakeRequest("DELETE", "/api/v1/hotels/"+s.hotelID, nil, s.ownerToken)
		s.Equal(http.StatusOK, res.Code)
		
		resList := s.MakeRequest("GET", "/api/v1/hotels", nil, s.ownerToken)
		s.NotContains(resList.Body.String(), s.hotelID)
	})
}

func TestLifecycleSuite(t *testing.T) {
	suite.Run(t, new(LifecycleSuite))
}
