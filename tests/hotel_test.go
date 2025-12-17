package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/ecelayes/pms-backend/internal/entity"
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
	
	var response entity.PaginatedResponse[entity.Hotel]
	json.Unmarshal(resList.Body.Bytes(), &response)
	s.NotEmpty(response.Data)
	found := false
	for _, h := range response.Data {
		if h.ID == hotelID {
			found = true
			break
		}
	}
	s.True(found, "Newly created hotel should be in the list")

	resDel := s.MakeRequest("DELETE", "/api/v1/hotels/"+hotelID, nil, s.token)
	s.Equal(http.StatusOK, resDel.Code)
}

func (s *HotelSuite) TestHotelValidation() {
	res := s.MakeRequest("POST", "/api/v1/hotels", map[string]string{
		"organization_id": s.orgID,
		"name":            "",
		"code":            "INV",
	}, s.token)
	s.Equal(http.StatusBadRequest, res.Code)

	res2 := s.MakeRequest("POST", "/api/v1/hotels", map[string]string{
		"organization_id": s.orgID,
		"name":            "Valid Name",
		"code":            "",
	}, s.token)
	s.Equal(http.StatusBadRequest, res2.Code)
}

func (s *HotelSuite) TestHotelNotFound() {
	res := s.MakeRequest("GET", "/api/v1/hotels/00000000-0000-0000-0000-000000000000", nil, s.token)
	s.Equal(http.StatusNotFound, res.Code)

	resUpd := s.MakeRequest("PUT", "/api/v1/hotels/00000000-0000-0000-0000-000000000000", map[string]string{
		"name": "Updated Name",
	}, s.token)
	s.Equal(http.StatusNotFound, resUpd.Code)

	resDel := s.MakeRequest("DELETE", "/api/v1/hotels/00000000-0000-0000-0000-000000000000", nil, s.token)
	s.Equal(http.StatusNotFound, resDel.Code)
}

func (s *HotelSuite) TestHotelDuplicateCode() {
	res := s.MakeRequest("POST", "/api/v1/hotels", map[string]string{
		"organization_id": s.orgID,
		"name":            "Unique Hotel",
		"code":            "UNI",
	}, s.token)
	s.Equal(http.StatusCreated, res.Code)

	res2 := s.MakeRequest("POST", "/api/v1/hotels", map[string]string{
		"organization_id": s.orgID,
		"name":            "Another Hotel",
		"code":            "UNI",
	}, s.token)
	s.Equal(http.StatusConflict, res2.Code)
}

func TestHotelSuite(t *testing.T) {
	suite.Run(t, new(HotelSuite))
}
