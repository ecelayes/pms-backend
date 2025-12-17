package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"github.com/stretchr/testify/suite"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type AvailabilitySuite struct {
	BaseSuite
	token   string
	orgID   string
	hotelID string
	roomID  string
}

func (s *AvailabilitySuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.token, s.orgID = s.GetAdminTokenAndOrg()

	resH := s.MakeRequest("POST", "/api/v1/hotels", map[string]string{
		"organization_id": s.orgID,
		"name":            "Avail Hotel",
		"code":            "AVH",
	}, s.token)
	var dataH map[string]string
	json.Unmarshal(resH.Body.Bytes(), &dataH)
	s.hotelID = dataH["hotel_id"]

	resR := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID,
		"name":           "Standard Room",
		"code":           "STD",
		"total_quantity": 10,
		"base_price":     100.0,
		"max_occupancy":  2, "max_adults": 2, "max_children": 0,
		"amenities":      []string{"wifi"},
	}, s.token)
	var dataR map[string]string
	json.Unmarshal(resR.Body.Bytes(), &dataR)
	s.roomID = dataR["room_type_id"]

	s.MakeRequest("POST", "/api/v1/pricing/bulk", map[string]interface{}{
		"room_type_id": s.roomID,
		"start":        "2025-06-01",
		"end":          "2025-06-10",
		"price":        150.0,
	}, s.token)

	s.MakeRequest("POST", "/api/v1/rate-plans", map[string]interface{}{
		"hotel_id":     s.hotelID,
		"room_type_id": s.roomID,
		"name":         "Standard Rate",
		"description":  "Standard Rate",
		"meal_plan": map[string]interface{}{
			"included":      false,
			"price_per_pax": 0,
			"type":          0,
		},
		"cancellation_policy": map[string]interface{}{
			"is_refundable": true,
			"rules":         []map[string]interface{}{},
		},
		"payment_policy": map[string]interface{}{
			"timing": 0, "method": 0,
		},
	}, s.token)
}

func (s *AvailabilitySuite) TestAvailabilitySearch() {
	url := "/api/v1/availability?hotel_id=" + s.hotelID + 
		"&start=2025-06-02&end=2025-06-05&adults=2&children=0&rooms=1"

	res := s.MakeRequest("GET", url, nil, "")
	
	s.Equal(http.StatusOK, res.Code)
	
	var response entity.PaginatedResponse[entity.AvailabilitySearch]
	err := json.Unmarshal(res.Body.Bytes(), &response)
	s.NoError(err)

	s.NotEmpty(response.Data, "Should return at least one room type")
	if len(response.Data) > 0 {
		s.Equal(s.roomID, response.Data[0].RoomTypeID)
	}
}

func (s *AvailabilitySuite) TestGlobalAvailabilitySearch() {
	url := "/api/v1/availability?start=2025-06-02&end=2025-06-05&adults=2&children=0&rooms=1"
	
	res := s.MakeRequest("GET", url, nil, "") 
	s.Equal(http.StatusOK, res.Code)

	var response entity.PaginatedResponse[entity.AvailabilitySearch]
	err := json.Unmarshal(res.Body.Bytes(), &response)
	s.NoError(err)
	
	found := false
	for _, result := range response.Data {
		if result.RoomTypeID == s.roomID {
			found = true
			break
		}
	}
	s.True(found, "Global search should return the room type from Avail Hotel")
}

func (s *AvailabilitySuite) TestAvailabilityPagination() {
	resR2 := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID,
		"name":           "Suite Room",
		"code":           "SUI",
		"total_quantity": 5,
		"base_price":     200.0,
		"max_occupancy":  4, "max_adults": 4, "max_children": 2,
		"amenities":      []string{"wifi", "jacuzzi"},
	}, s.token)
	var dataR2 map[string]string
	json.Unmarshal(resR2.Body.Bytes(), &dataR2)
	roomID2 := dataR2["room_type_id"]

	s.MakeRequest("POST", "/api/v1/pricing/bulk", map[string]interface{}{
		"room_type_id": roomID2,
		"start":        "2025-06-01",
		"end":          "2025-06-10",
		"price":        250.0,
	}, s.token)

	s.MakeRequest("POST", "/api/v1/rate-plans", map[string]interface{}{
		"hotel_id":     s.hotelID,
		"room_type_id": roomID2,
		"name":         "Suite Rate",
		"description":  "Suite Standard Rate",
		"meal_plan": map[string]interface{}{ "included": false, "type": 0, "price_per_pax": 0 },
		"cancellation_policy": map[string]interface{}{ "is_refundable": true, "rules": []map[string]interface{}{} },
		"payment_policy": map[string]interface{}{ "timing": 0, "method": 0 },
	}, s.token)

	url := "/api/v1/availability?start=2025-06-02&end=2025-06-05&adults=2&children=0&rooms=1&page=1&limit=1&hotel_id=" + s.hotelID
	res := s.MakeRequest("GET", url, nil, "")
	s.Equal(http.StatusOK, res.Code)

	var response entity.PaginatedResponse[entity.AvailabilitySearch]
	json.Unmarshal(res.Body.Bytes(), &response)

	s.Equal(1, len(response.Data))
	s.Equal(int64(2), response.Meta.TotalItems)
	s.Equal(2, response.Meta.TotalPages)
	s.Equal(1, response.Meta.Page)

	url2 := "/api/v1/availability?start=2025-06-02&end=2025-06-05&adults=2&children=0&rooms=1&page=2&limit=1&hotel_id=" + s.hotelID
	res2 := s.MakeRequest("GET", url2, nil, "")
	s.Equal(http.StatusOK, res2.Code)

	var response2 entity.PaginatedResponse[entity.AvailabilitySearch]
	json.Unmarshal(res2.Body.Bytes(), &response2)

	s.Equal(1, len(response2.Data))
	s.Equal(int64(2), response2.Meta.TotalItems)
	s.Equal(2, response2.Meta.Page)
	
	s.NotEqual(response.Data[0].RoomTypeID, response2.Data[0].RoomTypeID)
}

func (s *AvailabilitySuite) TestAvailabilitySoldOut() {
	resR := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID,
		"name":           "Limited Room", "code": "LTD",
		"total_quantity": 1,
		"base_price":     100.0,
		"max_occupancy":  2, "max_adults": 2, "max_children": 0,
		"amenities":      []string{"wifi"},
	}, s.token)
	var dataR map[string]string
	json.Unmarshal(resR.Body.Bytes(), &dataR)
	roomID := dataR["room_type_id"]

	s.MakeRequest("POST", "/api/v1/pricing/bulk", map[string]interface{}{
		"room_type_id": roomID,
		"start":        "2025-08-01", "end": "2025-08-05",
		"price":        100.0,
	}, s.token)

	s.MakeRequest("POST", "/api/v1/rate-plans", map[string]interface{}{
		"hotel_id":     s.hotelID, "room_type_id": roomID,
		"name": "Standard Rate", 
		"meal_plan": map[string]interface{}{ "included": false, "type": 0, "price_per_pax": 0 },
		"cancellation_policy": map[string]interface{}{ "is_refundable": true, "rules": []map[string]interface{}{} },
		"payment_policy": map[string]interface{}{ "timing": 0, "method": 0 },
	}, s.token)

	resRes := s.MakeRequest("POST", "/api/v1/reservations", map[string]interface{}{
		"room_type_id":     roomID,
		"guest_email":      "soldout@test.com",
		"guest_first_name": "Sold", "guest_last_name": "Out",
		"start":            "2025-08-01", "end": "2025-08-03",
		"adults":           1, "children": 0,
	}, "")
	s.Equal(http.StatusCreated, resRes.Code)

	url := "/api/v1/availability?hotel_id=" + s.hotelID + "&start=2025-08-01&end=2025-08-03&adults=1"
	res := s.MakeRequest("GET", url, nil, "")
	s.Equal(http.StatusOK, res.Code)

	var response entity.PaginatedResponse[entity.AvailabilitySearch]
	json.Unmarshal(res.Body.Bytes(), &response)

	for _, result := range response.Data {
		s.NotEqual(roomID, result.RoomTypeID, "Sold out room should not be available")
	}

	url2 := "/api/v1/availability?hotel_id=" + s.hotelID + "&start=2025-08-03&end=2025-08-05&adults=1"
	res2 := s.MakeRequest("GET", url2, nil, "")
	
	var response2 entity.PaginatedResponse[entity.AvailabilitySearch]
	json.Unmarshal(res2.Body.Bytes(), &response2)
	
	found := false
	for _, result := range response2.Data {
		if result.RoomTypeID == roomID {
			found = true
			break
		}
	}
	s.True(found, "Room should be available after existing reservation check-out")
}

func (s *AvailabilitySuite) TestAvailabilityOccupancy() {
	url := "/api/v1/availability?hotel_id=" + s.hotelID + "&start=2025-06-02&end=2025-06-05&adults=3"
	res := s.MakeRequest("GET", url, nil, "")
	s.Equal(http.StatusOK, res.Code)

	var response entity.PaginatedResponse[entity.AvailabilitySearch]
	json.Unmarshal(res.Body.Bytes(), &response)

	for _, result := range response.Data {
		s.NotEqual(s.roomID, result.RoomTypeID, "Room with max_occupancy 2 should not show for 3 adults")
	}
}

func (s *AvailabilitySuite) TestAvailabilityMissingPrice() {
	resR := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID,
		"name":           "No Price Room", "code": "NPR",
		"total_quantity": 5,
		"base_price":     0.0,
		"max_occupancy":  2, "max_adults": 2, "max_children": 0,
		"amenities":      []string{"wifi"},
	}, s.token)
	var dataR map[string]string
	json.Unmarshal(resR.Body.Bytes(), &dataR)
	roomID := dataR["room_type_id"]

	s.MakeRequest("POST", "/api/v1/rate-plans", map[string]interface{}{
		"hotel_id":     s.hotelID, "room_type_id": roomID,
		"name": "Standard Rate", 
		"meal_plan": map[string]interface{}{ "included": false, "type": 0, "price_per_pax": 0 },
		"cancellation_policy": map[string]interface{}{ "is_refundable": true, "rules": []map[string]interface{}{} },
		"payment_policy": map[string]interface{}{ "timing": 0, "method": 0 },
	}, s.token)

	url := "/api/v1/availability?hotel_id=" + s.hotelID + "&start=2025-09-01&end=2025-09-03&adults=2"
	res := s.MakeRequest("GET", url, nil, "")
	s.Equal(http.StatusOK, res.Code)

	var response entity.PaginatedResponse[entity.AvailabilitySearch]
	json.Unmarshal(res.Body.Bytes(), &response)

	for _, result := range response.Data {
		s.NotEqual(roomID, result.RoomTypeID, "Room without pricing (and 0 base price) should not be available")
	}
}

func (s *AvailabilitySuite) TestAvailabilityDateValidation() {
	url := "/api/v1/availability?hotel_id=" + s.hotelID + "&start=2025-06-05&end=2025-06-02&adults=2"
	res := s.MakeRequest("GET", url, nil, "")
	
	s.Equal(http.StatusBadRequest, res.Code)
}

func TestAvailabilitySuite(t *testing.T) {
	suite.Run(t, new(AvailabilitySuite))
}
