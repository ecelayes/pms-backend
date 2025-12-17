package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type PricingSuite struct {
	BaseSuite
	token   string
	orgID   string
	hotelID string
	roomID  string
}

func (s *PricingSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.token, s.orgID = s.GetAdminTokenAndOrg()

	resH := s.MakeRequest("POST", "/api/v1/hotels", map[string]string{
		"organization_id": s.orgID,
		"name":            "Price Hotel", "code": "PHO",
	}, s.token)
	var dataH map[string]string
	json.Unmarshal(resH.Body.Bytes(), &dataH)
	s.hotelID = dataH["hotel_id"]
	
	resR := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       s.hotelID, 
		"name":           "R", "code": "RRR", 
		"total_quantity": 5,
		"max_occupancy":  2, "max_adults": 2, "max_children": 0,
		"amenities":      []string{"wifi"},
	}, s.token)
	
	var dataR map[string]string
	json.Unmarshal(resR.Body.Bytes(), &dataR)
	s.roomID = dataR["room_type_id"]
}

func (s *PricingSuite) TestBulkPricingLogic() {
	res := s.MakeRequest("POST", "/api/v1/pricing/bulk", map[string]interface{}{
		"room_type_id": s.roomID,
		"start": "2025-01-01", "end": "2025-01-31", 
		"price": 100.0,
	}, s.token)
	s.Equal(http.StatusOK, res.Code)

	res2 := s.MakeRequest("POST", "/api/v1/pricing/bulk", map[string]interface{}{
		"room_type_id": s.roomID,
		"start": "2025-01-10", "end": "2025-01-15", 
		"price": 200.0,
	}, s.token)
	s.Equal(http.StatusOK, res2.Code)

	resGet := s.MakeRequest("GET", "/api/v1/pricing/rules?room_type_id="+s.roomID, nil, s.token)
	s.Equal(http.StatusOK, resGet.Code)

	var response entity.PaginatedResponse[entity.PriceRule]
	json.Unmarshal(resGet.Body.Bytes(), &response)
	rules := response.Data

	s.Len(rules, 3, "I should have cut the base ruler into three pieces.")
	
	if len(rules) == 3 {
		s.Equal(100.0, rules[0].Price, "Fragment 1 incorrect")
		s.Equal(200.0, rules[1].Price, "Fragment 2 incorrect")
		s.Equal(100.0, rules[2].Price, "Fragment 3 incorrect")
	}
	s.Equal(3, int(response.Meta.TotalItems))
}

func (s *PricingSuite) TestListByHotel() {
	s.MakeRequest("POST", "/api/v1/pricing/bulk", map[string]interface{}{
		"room_type_id": s.roomID,
		"start": "2025-03-01", "end": "2025-03-05", 
		"price": 50.0,
	}, s.token)

	resGet := s.MakeRequest("GET", "/api/v1/pricing/rules?hotel_id="+s.hotelID, nil, s.token)
	s.Equal(http.StatusOK, resGet.Code)
	
	var response entity.PaginatedResponse[entity.PriceRule]
	json.Unmarshal(resGet.Body.Bytes(), &response)
	rules := response.Data

	s.Len(rules, 1)
	if len(rules) > 0 {
		s.Equal(50.0, rules[0].Price)
		s.Equal(s.roomID, rules[0].RoomTypeID)
	}
}

func (s *PricingSuite) TestDeletePriceRule() {
	s.MakeRequest("POST", "/api/v1/pricing/bulk", map[string]interface{}{
		"room_type_id": s.roomID,
		"start": "2025-02-01", "end": "2025-02-10", 
		"price": 150.0,
	}, s.token)

	resGet := s.MakeRequest("GET", "/api/v1/pricing/rules?room_type_id="+s.roomID, nil, s.token)
	var response entity.PaginatedResponse[entity.PriceRule]
	json.Unmarshal(resGet.Body.Bytes(), &response)
	rules := response.Data
	ruleID := rules[0].ID

	resDel := s.MakeRequest("DELETE", "/api/v1/pricing/rules/"+ruleID, nil, s.token)
	s.Equal(http.StatusOK, resDel.Code)

	resGet2 := s.MakeRequest("GET", "/api/v1/pricing/rules?room_type_id="+s.roomID, nil, s.token)
	var response2 entity.PaginatedResponse[entity.PriceRule]
	json.Unmarshal(resGet2.Body.Bytes(), &response2)
	rules2 := response2.Data
	s.Len(rules2, 0, "The rule should have been deleted.")
}

func (s *PricingSuite) TestPricingValidation() {
	res := s.MakeRequest("POST", "/api/v1/pricing/bulk", map[string]interface{}{
		"room_type_id": s.roomID,
		"start":        "2025-02-01",
		"end":          "2025-01-01",
		"price":        100.0,
	}, s.token)
	s.Equal(http.StatusBadRequest, res.Code)

	res2 := s.MakeRequest("POST", "/api/v1/pricing/bulk", map[string]interface{}{
		"room_type_id": s.roomID,
		"start":        "2025-02-01",
		"end":          "2025-02-05",
		"price":        -50.0,
	}, s.token)
	s.Equal(http.StatusBadRequest, res2.Code)
}

func (s *PricingSuite) TestPricingNotFound() {
	resDel := s.MakeRequest("DELETE", "/api/v1/pricing/rules/00000000-0000-0000-0000-000000000000", nil, s.token)
	s.Equal(http.StatusNotFound, resDel.Code)
}

func TestPricingSuite(t *testing.T) {
	suite.Run(t, new(PricingSuite))
}
