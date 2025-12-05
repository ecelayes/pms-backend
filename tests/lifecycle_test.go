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
	s.Run("1. Auth Setup", func() {
		s.MakeRequest("POST", "/api/v1/auth/register", map[string]string{"email": "ceo@chain.com", "password": "Pass"}, "")
		res := s.MakeRequest("POST", "/api/v1/auth/login", map[string]string{"email": "ceo@chain.com", "password": "Pass"}, "")
		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.ownerToken = data["token"]
	})

	s.Run("2. Create Hotel", func() {
		res := s.MakeRequest("POST", "/api/v1/hotels", map[string]string{"name": "Grand Lifecycle", "code": "LIF"}, s.ownerToken)
		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.hotelID = data["hotel_id"]
	})

	s.Run("3. Create Family Room", func() {
		res := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
			"hotel_id":       s.hotelID,
			"name":           "Family Suite",
			"code":           "FAM",
			"total_quantity": 5,
			"max_occupancy":  4,
			"max_adults":     2,
			"max_children":   2,
			"amenities":      []string{"wifi", "crib"},
		}, s.ownerToken)
		s.Equal(http.StatusCreated, res.Code)
		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.roomID = data["room_type_id"]
	})

	s.Run("4. Set Pricing", func() {
		s.MakeRequest("POST", "/api/v1/pricing/rules", map[string]interface{}{
			"room_type_id": s.roomID,
			"start": "2025-10-01", "end": "2025-10-05", "price": 200.0, "priority": 10,
		}, s.ownerToken)
	})

	s.Run("5. Availability Logic", func() {
		resOK := s.MakeRequest("GET", "/api/v1/availability?start=2025-10-01&end=2025-10-05&adults=2&children=2", nil, "")
		s.Contains(resOK.Body.String(), `"available_qty":5`)

		resFail := s.MakeRequest("GET", "/api/v1/availability?start=2025-10-01&end=2025-10-05&adults=3", nil, "")
		s.NotContains(resFail.Body.String(), s.roomID)

		resMulti := s.MakeRequest("GET", "/api/v1/availability?start=2025-10-01&end=2025-10-05&adults=4&rooms=2", nil, "")
		s.Contains(resMulti.Body.String(), `"total_price":1600`)
	})

	s.Run("6. Customer Reserves", func() {
		res := s.MakeRequest("POST", "/api/v1/reservations", map[string]interface{}{
			"room_type_id":     s.roomID,
			"guest_email":      "family@vacation.com",
			"guest_first_name": "Pepito",
			"guest_last_name":  "Perez",
			"start":            "2025-10-01",
			"end":              "2025-10-05",
			"adults":           2,
			"children":         2,
		}, "")
		s.Equal(http.StatusCreated, res.Code)

		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.resCode = data["reservation_code"]
	})

	s.Run("7. Update Guest on Reserve", func() {
		res := s.MakeRequest("POST", "/api/v1/reservations", map[string]interface{}{
			"room_type_id":     s.roomID,
			"guest_email":      "family@vacation.com",
			"guest_first_name": "Pepito Updated",
			"guest_last_name":  "Perez",
			"guest_phone":      "99999",
			"start":            "2025-10-01",
			"end":              "2025-10-05",
			"adults":           2,
			"children":         0,
		}, "")
		s.Equal(http.StatusCreated, res.Code)
	})
}

func TestLifecycleSuite(t *testing.T) {
	suite.Run(t, new(LifecycleSuite))
}
