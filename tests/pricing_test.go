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

	var rules []entity.PriceRule
	json.Unmarshal(resGet.Body.Bytes(), &rules)

	s.Len(rules, 3, "Debería haber cortado la regla base en 3 fragmentos")
	
	if len(rules) == 3 {
		s.Equal(100.0, rules[0].Price, "Fragmento 1 incorrecto")
		s.Equal(200.0, rules[1].Price, "Fragmento 2 (nuevo) incorrecto")
		s.Equal(100.0, rules[2].Price, "Fragmento 3 incorrecto")
	}
}

func TestPricingSuite(t *testing.T) {
	suite.Run(t, new(PricingSuite))
}
