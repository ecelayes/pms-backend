package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LifecycleSuite struct {
	BaseSuite
	superToken string
	ownerToken string
	orgID      string
	hotelID    string
	roomID     string
}

func (s *LifecycleSuite) TestFullLifecycle() {
	s.superToken = s.GetSuperAdminToken()

	s.Run("1. Create Organization", func() {
		res := s.MakeRequest("POST", "/api/v1/organizations", map[string]string{
			"name": "Global Hotels Corp",
		}, s.superToken)
		s.Equal(http.StatusCreated, res.Code)
		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.orgID = data["organization_id"]
	})

	s.Run("2. Create Owner User", func() {
		res := s.MakeRequest("POST", "/api/v1/users", map[string]string{
			"organization_id": s.orgID,
			"email":           "ceo@global.com",
			"password":        "pass",
			"role":            "owner",
		}, s.superToken)
		s.Equal(http.StatusCreated, res.Code)
	})

	s.Run("3. Login Owner", func() {
		res := s.MakeRequest("POST", "/api/v1/auth/login", map[string]string{
			"email":    "ceo@global.com",
			"password": "pass",
		}, "")
		s.Equal(http.StatusOK, res.Code)
		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.ownerToken = data["token"]
	})

	s.Run("4. Create Hotel", func() {
		res := s.MakeRequest("POST", "/api/v1/hotels", map[string]string{
			"organization_id": s.orgID,
			"name":            "Grand Global",
			"code":            "GLB",
		}, s.ownerToken)
		
		s.Equal(http.StatusCreated, res.Code)
		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.hotelID = data["hotel_id"]
	})

	s.Run("5. Create Room", func() {
		res := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
			"hotel_id":       s.hotelID,
			"name":           "Suite",
			"code":           "SUI",
			"total_quantity": 10,
			"max_occupancy":  2, "max_adults": 2, "max_children": 0,
			"amenities":      []string{"wifi"},
		}, s.ownerToken)
		
		s.Equal(http.StatusCreated, res.Code)
		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.roomID = data["room_type_id"]
	})

	s.Run("6. Set Pricing", func() {
		res := s.MakeRequest("POST", "/api/v1/pricing/rules", map[string]interface{}{
			"room_type_id": s.roomID,
			"start": "2025-10-01", "end": "2025-10-10", "price": 200.0, "priority": 10,
		}, s.ownerToken)
		s.Equal(http.StatusCreated, res.Code)
	})

	s.Run("7. Customer Reservation", func() {
		res := s.MakeRequest("POST", "/api/v1/reservations", map[string]interface{}{
			"room_type_id":     s.roomID,
			"guest_email":      "client@mail.com",
			"guest_first_name": "Client", "guest_last_name": "One",
			"start":            "2025-10-01", "end": "2025-10-05",
			"adults":           2, "children": 0,
		}, "")
		
		if res.Code != http.StatusCreated {
			s.T().Logf("Reservation failed: %s", res.Body.String())
		}
		s.Equal(http.StatusCreated, res.Code)
	})
}

func TestLifecycleSuite(t *testing.T) {
	suite.Run(t, new(LifecycleSuite))
}
