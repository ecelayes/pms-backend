package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type HotelSuite struct {
	BaseSuite
	token string
	orgID string
}

func (s *HotelSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.token, s.orgID = s.GetAdminTokenAndOrg()
}

func (s *HotelSuite) TestCRUDHotel() {
	res := s.MakeRequest("POST", "/api/v1/hotels", map[string]string{
		"organization_id": s.orgID,
		"name":            "Hotel Test",
		"code":            "HTL",
	}, s.token)
	s.Require().Equal(http.StatusCreated, res.Code)

	var data map[string]string
	json.Unmarshal(res.Body.Bytes(), &data)
	hotelID := data["hotel_id"]

	resList := s.MakeRequest("GET", "/api/v1/hotels?organization_id="+s.orgID, nil, s.token)
	s.Equal(http.StatusOK, resList.Code)
	s.Contains(resList.Body.String(), "Hotel Test")

	resDel := s.MakeRequest("DELETE", "/api/v1/hotels/"+hotelID, nil, s.token)
	s.Equal(http.StatusOK, resDel.Code)
}

func TestHotelSuite(t *testing.T) {
	suite.Run(t, new(HotelSuite))
}
