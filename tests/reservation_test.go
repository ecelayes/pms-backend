package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
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
		"base_price":     100.0,
		"max_occupancy":  4, "max_adults": 2, "max_children": 2,
		"amenities":      []string{"wifi"},
	}, s.token)
	s.Require().Equal(http.StatusCreated, resR.Code)

	var dataR map[string]string
	json.Unmarshal(resR.Body.Bytes(), &dataR)
	s.roomTypeID = dataR["room_type_id"]

	s.MakeRequest("POST", "/api/v1/pricing/bulk", map[string]interface{}{
		"room_type_id": s.roomTypeID,
		"start": "2025-01-01", "end": "2025-01-10", 
		"price": 100.0,
	}, s.token)
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
			"type":          1, 
		},
		"cancellation_policy": map[string]interface{}{"is_refundable": true, "rules": []interface{}{}},
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

func (s *ReservationSuite) TestReservationFallbackPrice() {
	resR := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID, 
		"name":           "Fallback Room", "code": "FBK", 
		"total_quantity": 5,
		"base_price":     120.0,
		"max_occupancy":  2, "max_adults": 2, "max_children": 0,
		"amenities":      []string{"wifi"},
	}, s.token)
	s.Require().Equal(http.StatusCreated, resR.Code)

	var dataR map[string]string
	json.Unmarshal(resR.Body.Bytes(), &dataR)
	roomID := dataR["room_type_id"]

	res := s.MakeRequest("POST", "/api/v1/reservations", map[string]interface{}{
		"room_type_id":     roomID,
		"guest_email":      "fallback@test.com",
		"guest_first_name": "Fall", "guest_last_name": "Back",
		"start":            "2026-05-01", "end": "2026-05-03",
		"adults":           2, "children": 0,
	}, "")
	
	s.Equal(http.StatusCreated, res.Code, "La reserva debería crearse usando el precio base")

	var dataRes map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &dataRes)
	code := dataRes["reservation_code"].(string)

	resGet := s.MakeRequest("GET", "/api/v1/reservations/"+code, nil, "")
	
	var resData entity.Reservation
	json.Unmarshal(resGet.Body.Bytes(), &resData)

	s.Equal(240.0, resData.TotalPrice)
}

func (s *ReservationSuite) TestCancellationPenalty() {
	s.MakeRequest("POST", "/api/v1/pricing/bulk", map[string]interface{}{
		"room_type_id": s.roomTypeID,
		"start":        "2026-06-01",
		"end":          "2026-06-10",
		"price":        100.0,
	}, s.token)

	resRP := s.MakeRequest("POST", "/api/v1/rate-plans", map[string]interface{}{
		"hotel_id":     s.hotelID,
		"room_type_id": s.roomTypeID,
		"name":         "Strict 50",
		"meal_plan":    map[string]interface{}{"included": false},
		"cancellation_policy": map[string]interface{}{
			"is_refundable": true,
			"rules": []map[string]interface{}{
				{
					"hours_before_check_in": 10000,
					"penalty_type":          1,
					"penalty_value":         50.0,
				},
			},
		},
		"payment_policy": map[string]interface{}{"timing": 0},
	}, s.token)
	s.Require().Equal(http.StatusCreated, resRP.Code)
	
	var dataRP map[string]string
	json.Unmarshal(resRP.Body.Bytes(), &dataRP)
	planID := dataRP["rate_plan_id"]

	res := s.MakeRequest("POST", "/api/v1/reservations", map[string]interface{}{
		"room_type_id":     s.roomTypeID,
		"rate_plan_id":     planID,
		"guest_email":      "penalty@test.com",
		"guest_first_name": "Pen", "guest_last_name": "Alty",
		"start":            "2026-06-01", "end": "2026-06-03",
		"adults":           2, "children": 0,
	}, "")
	s.Require().Equal(http.StatusCreated, res.Code)

	var dataRes map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &dataRes)
	resCode := dataRes["reservation_code"].(string)

	resGet := s.MakeRequest("GET", "/api/v1/reservations/"+resCode, nil, "")
	var resData entity.Reservation
	json.Unmarshal(resGet.Body.Bytes(), &resData)
	resID := resData.ID

	resPreview := s.MakeRequest("GET", "/api/v1/reservations/"+resID+"/cancel-preview", nil, s.token)
	s.Equal(http.StatusOK, resPreview.Code)

	var previewData map[string]interface{}
	json.Unmarshal(resPreview.Body.Bytes(), &previewData)

	s.Equal(100.0, previewData["penalty_amount"])
}

func (s *ReservationSuite) TestConcurrencyOverbooking() {
	resR := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID,
		"name":           "Single Room", "code": "SGL",
		"total_quantity": 1,
		"base_price":     100.0,
		"max_occupancy":  2, "max_adults": 2, "max_children": 0,
		"amenities":      []string{"wifi"},
	}, s.token)
	s.Require().Equal(http.StatusCreated, resR.Code)

	var dataR map[string]string
	json.Unmarshal(resR.Body.Bytes(), &dataR)
	targetRoomID := dataR["room_type_id"]

	s.MakeRequest("POST", "/api/v1/pricing/bulk", map[string]interface{}{
		"room_type_id": targetRoomID,
		"start": "2026-12-01", "end": "2026-12-05", 
		"price": 100.0,
	}, s.token)

	
	guestEmail := "concurrent@test.com"
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO guests (email, first_name, last_name, phone, created_at, updated_at) 
		VALUES ($1, 'Pre', 'Created', '555-5555', NOW(), NOW())
	`, guestEmail)
	s.Require().NoError(err)

	concurrentReqs := 10
	var wg sync.WaitGroup
	wg.Add(concurrentReqs)

	successCount := 0
	failCount := 0
	var mu sync.Mutex

	for i := 0; i < concurrentReqs; i++ {
		go func(idx int) {
			defer wg.Done()
			
			payload := map[string]interface{}{
				"room_type_id":     targetRoomID,
				"guest_email":      guestEmail,
				"guest_first_name": "Race", "guest_last_name": "Condition",
				"start":            "2026-12-01", "end": "2026-12-02",
				"adults":           1, "children": 0,
			}

			res := s.MakeRequest("POST", "/api/v1/reservations", payload, "")
			
			mu.Lock()
			if res.Code == http.StatusCreated {
				successCount++
			} else if res.Code == http.StatusConflict {
				failCount++
			} else {
				s.T().Logf("Unexpected status: %d, body: %s", res.Code, res.Body.String())
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	s.Equal(1, successCount, "Only one reservation should be successful.")
	s.Equal(concurrentReqs-1, failCount, "The rest should fail due to overbooking (409)")
}

func TestReservationSuite(t *testing.T) {
	suite.Run(t, new(ReservationSuite))
}
