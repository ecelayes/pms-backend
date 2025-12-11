package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RatePlanSuite struct {
	BaseSuite
	token      string
	orgID      string
	hotelID    string
	roomTypeID string
}

func (s *RatePlanSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.token, s.orgID = s.GetAdminTokenAndOrg()

	resH := s.MakeRequest("POST", "/api/v1/hotels", map[string]string{
		"organization_id": s.orgID,
		"name":            "RatePlan Hotel",
		"code":            "RPH",
	}, s.token)
	s.Require().Equal(http.StatusCreated, resH.Code)
	
	var dataH map[string]string
	json.Unmarshal(resH.Body.Bytes(), &dataH)
	s.hotelID = dataH["hotel_id"]

	resR := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID,
		"name":           "Deluxe",
		"code":           "DLX",
		"total_quantity": 10,
		"max_occupancy":  2, "max_adults": 2, "max_children": 0,
		"amenities":      []string{"wifi"},
	}, s.token)
	s.Require().Equal(http.StatusCreated, resR.Code)
	
	var dataR map[string]string
	json.Unmarshal(resR.Body.Bytes(), &dataR)
	s.roomTypeID = dataR["room_type_id"]

	s.MakeRequest("POST", "/api/v1/pricing/rules", map[string]interface{}{
		"room_type_id": s.roomTypeID,
		"start": "2026-01-01", "end": "2026-01-31",
        "price": 100.0, "priority": 1,
	}, s.token)
}

func (s *RatePlanSuite) TestRatePlanLifecycle() {
	reqBody := map[string]interface{}{
		"hotel_id":     s.hotelID,
		"room_type_id": s.roomTypeID,
		"name":         "Breakfast Included",
		"description":  "Bed and Breakfast",
		"meal_plan": map[string]interface{}{
			"included":      true,
			"price_per_pax": 15.0,
			"type":          1, 
		},
		"cancellation_policy": map[string]interface{}{
			"is_refundable": true,
			"rules": []map[string]interface{}{
				{"hours_before_check_in": 48, "penalty_type": 1, "penalty_value": 100},
			},
		},
		"payment_policy": map[string]interface{}{
			"timing": 0, "method": 0,
		},
	}

	res := s.MakeRequest("POST", "/api/v1/rate-plans", reqBody, s.token)
	s.Require().Equal(http.StatusCreated, res.Code)
	
	var data map[string]string
	json.Unmarshal(res.Body.Bytes(), &data)
	planID := data["rate_plan_id"]

	resList := s.MakeRequest("GET", "/api/v1/rate-plans?hotel_id="+s.hotelID, nil, s.token)
	s.Equal(http.StatusOK, resList.Code)
	s.Contains(resList.Body.String(), "Breakfast Included")

	updateBody := map[string]interface{}{
		"name": "Breakfast & Dinner",
		"meal_plan": map[string]interface{}{
			"included":      true,
			"price_per_pax": 30.0,
		},
	}
	resUpdate := s.MakeRequest("PUT", "/api/v1/rate-plans/"+planID, updateBody, s.token)
	s.Equal(http.StatusOK, resUpdate.Code)

	resRes := s.MakeRequest("POST", "/api/v1/reservations", map[string]interface{}{
		"room_type_id":     s.roomTypeID,
		"rate_plan_id":     planID,
		"guest_email":      "check@integrity.com",
		"guest_first_name": "Integrity", "guest_last_name": "Check",
		"start":            "2026-01-01", "end": "2026-01-02",
		"adults":           2, "children": 0,
	}, "")
	s.Require().Equal(http.StatusCreated, resRes.Code)

	resDelFail := s.MakeRequest("DELETE", "/api/v1/rate-plans/"+planID, nil, s.token)
	s.Equal(http.StatusInternalServerError, resDelFail.Code)
	s.Contains(resDelFail.Body.String(), "active reservations depend on it")

	var resData map[string]string
	json.Unmarshal(resRes.Body.Bytes(), &resData)
	resCode := resData["reservation_code"]
	
	resGet := s.MakeRequest("GET", "/api/v1/reservations/"+resCode, nil, "")
	var resObj map[string]interface{}
	json.Unmarshal(resGet.Body.Bytes(), &resObj)
	resID := resObj["id"].(string)

	s.MakeRequest("POST", "/api/v1/reservations/"+resID+"/cancel", nil, "")

	resDelSuccess := s.MakeRequest("DELETE", "/api/v1/rate-plans/"+planID, nil, s.token)
	s.Equal(http.StatusOK, resDelSuccess.Code)
}

func TestRatePlanSuite(t *testing.T) {
	suite.Run(t, new(RatePlanSuite))
}
