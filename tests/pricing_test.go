package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PricingSuite struct {
	BaseSuite
	token  string
	orgID  string
	roomID string
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
	
	resR := s.MakeRequest("POST", "/api/v1/room-types", map[string]interface{}{
		"hotel_id":       dataH["hotel_id"], 
		"name":           "R", "code": "RRR", 
		"total_quantity": 5,
		"max_occupancy":  2, "max_adults": 2, "max_children": 0,
		"amenities":      []string{"wifi"},
	}, s.token)
	
	var dataR map[string]string
	json.Unmarshal(resR.Body.Bytes(), &dataR)
	s.roomID = dataR["room_type_id"]
}

func (s *PricingSuite) TestCRUDPricing() {
	res := s.MakeRequest("POST", "/api/v1/pricing/rules", map[string]interface{}{
		"room_type_id": s.roomID,
		"start": "2025-01-01", "end": "2025-01-10", "price": 100.0, "priority": 0,
	}, s.token)
	s.Equal(http.StatusCreated, res.Code)
}

func TestPricingSuite(t *testing.T) {
	suite.Run(t, new(PricingSuite))
}
