package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type LifecycleSuite struct {
	BaseSuite
	superToken string
	ownerToken string
	orgID      string
	propertyID string
	unitTypeID string
	ratePlanID string
}

func (s *LifecycleSuite) TestFullLifecycle() {
	s.superToken = s.GetSuperAdminToken()

	s.Run("1. Create Organization", func() {
		res := s.MakeRequest("POST", "/api/v1/organizations", map[string]string{
			"name": "Global Properties Corp",
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
			"first_name":      "The",
			"last_name":       "CEO",
			"phone":           "111-2222",
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

	s.Run("4. Create Property", func() {
		res := s.MakeRequest("POST", "/api/v1/properties", map[string]string{
			"organization_id": s.orgID,
			"name":            "Grand Global",
			"code":            "GLB",
			"type":            "HOTEL",
		}, s.ownerToken)
		
		s.Equal(http.StatusCreated, res.Code)
		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.propertyID = data["property_id"]
	})

	s.Run("5. Create UnitType", func() {
		res := s.MakeRequest("POST", "/api/v1/unit-types", map[string]interface{}{
			"property_id":    s.propertyID,
			"name":           "Suite",
			"code":           "SUI",
			"total_quantity": 10,
			"max_occupancy":  2, "max_adults": 2, "max_children": 0,
			"amenities":      []string{"wifi"},
		}, s.ownerToken)
		
		s.Equal(http.StatusCreated, res.Code)
		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.unitTypeID = data["unit_type_id"]
	})

	s.Run("6. Set Pricing (Base Rate)", func() {
		res := s.MakeRequest("POST", "/api/v1/pricing/bulk", map[string]interface{}{
			"unit_type_id": s.unitTypeID,
			"start": "2025-10-01", "end": "2025-10-10", 
			"price": 200.0,
		}, s.ownerToken)
		s.Equal(http.StatusOK, res.Code)
	})

	s.Run("7. Create Rate Plan (Bed & Breakfast)", func() {
		reqBody := map[string]interface{}{
			"property_id":  s.propertyID,
			"unit_type_id": s.unitTypeID,
			"name":         "Bed & Breakfast",
			"description":  "Standard rate with breakfast included",
			"meal_plan": map[string]interface{}{
				"included":      true,
				"price_per_pax": 25.0,
				"type":          1, 
			},
			"cancellation_policy": map[string]interface{}{
				"is_refundable": true,
				"rules": []interface{}{}, 
			},
			"payment_policy": map[string]interface{}{
				"timing": 0, "method": 0,
			},
		}

		res := s.MakeRequest("POST", "/api/v1/rate-plans", reqBody, s.ownerToken)
		s.Require().Equal(http.StatusCreated, res.Code)
		
		var data map[string]string
		json.Unmarshal(res.Body.Bytes(), &data)
		s.ratePlanID = data["rate_plan_id"]
	})

	s.Run("8. Customer Reservation with Rate Plan", func() {
		res := s.MakeRequest("POST", "/api/v1/reservations", map[string]interface{}{
			"unit_type_id":     s.unitTypeID,
			"rate_plan_id":     s.ratePlanID,
			"guest_email":      "client@mail.com",
			"guest_first_name": "Client", "guest_last_name": "One",
			"start":            "2025-10-01", "end": "2025-10-05",
			"adults":           2, "children": 0,
		}, "")
		
		if res.Code != http.StatusCreated {
			s.T().Logf("Reservation failed: %s", res.Body.String())
		}
		s.Require().Equal(http.StatusCreated, res.Code)

		var dataRes map[string]interface{}
		json.Unmarshal(res.Body.Bytes(), &dataRes)
		code := dataRes["reservation_code"].(string)

		resGet := s.MakeRequest("GET", "/api/v1/reservations/"+code, nil, "")
		s.Equal(http.StatusOK, resGet.Code)

		var reservation entity.Reservation
		json.Unmarshal(resGet.Body.Bytes(), &reservation)

		s.Equal(1000.0, reservation.TotalPrice, "El precio total debe incluir alojamiento + desayuno")
		s.Equal(s.ratePlanID, *reservation.RatePlanID)
	})
}

func TestLifecycleSuite(t *testing.T) {
	suite.Run(t, new(LifecycleSuite))
}
