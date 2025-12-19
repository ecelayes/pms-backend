package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type UnitTypeSuite struct {
	BaseSuite
	token      string
	orgID      string
	propertyID string
}

func (s *UnitTypeSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.token, s.orgID = s.GetAdminTokenAndOrg()

	res := s.MakeRequest("POST", "/api/v1/properties", map[string]string{
		"organization_id": s.orgID,
		"name":            "UnitType Property",
		"code":            "UPR",
		"type":            "HOTEL",
	}, s.token)
	s.Require().Equal(http.StatusCreated, res.Code)
	
	var data map[string]string
	json.Unmarshal(res.Body.Bytes(), &data)
	s.propertyID = data["property_id"]
}

func (s *UnitTypeSuite) TestCRUDUnitType() {
	res := s.MakeRequest("POST", "/api/v1/unit-types", map[string]interface{}{
		"property_id":    s.propertyID,
		"name":           "Suite",
		"code":           "SUI",
		"total_quantity": 5,
		"base_price":     150.0,
		"max_occupancy":  4,
		"max_adults":     2,
		"max_children":   2,
		"amenities":      []string{"wifi"},
	}, s.token)
	s.Equal(http.StatusCreated, res.Code)
	
	var data map[string]string
	json.Unmarshal(res.Body.Bytes(), &data)
	id := data["unit_type_id"]

	resGet := s.MakeRequest("GET", "/api/v1/unit-types/"+id, nil, s.token)
	s.Equal(http.StatusOK, resGet.Code)
	
	var unitTypeMap map[string]interface{}
	json.Unmarshal(resGet.Body.Bytes(), &unitTypeMap)
	s.Equal(150.0, unitTypeMap["base_price"])

	resUpd := s.MakeRequest("PUT", "/api/v1/unit-types/"+id, map[string]interface{}{
		"base_price": 200.0,
	}, s.token)
	s.Equal(http.StatusOK, resUpd.Code)

	resGet2 := s.MakeRequest("GET", "/api/v1/unit-types/"+id, nil, s.token)
	json.Unmarshal(resGet2.Body.Bytes(), &unitTypeMap)
	s.Equal(200.0, unitTypeMap["base_price"])
}

func (s *UnitTypeSuite) TestListUnitTypes() {
	s.MakeRequest("POST", "/api/v1/unit-types", map[string]interface{}{
		"property_id":    s.propertyID,
		"name":           "List Test UnitType",
		"code":           "LTU",
		"total_quantity": 10,
		"base_price":     100.0,
		"max_occupancy":  2, "max_adults": 2, "max_children": 0,
		"amenities":      []string{"wifi"},
	}, s.token)

	res := s.MakeRequest("GET", "/api/v1/unit-types?property_id="+s.propertyID+"&page=1&limit=5", nil, s.token)
	s.Equal(http.StatusOK, res.Code)

	var response entity.PaginatedResponse[entity.UnitType]
	json.Unmarshal(res.Body.Bytes(), &response)

	s.NotEmpty(response.Data)
	s.Equal(1, response.Meta.Page)
	s.Equal(5, response.Meta.Limit)
	s.GreaterOrEqual(response.Meta.TotalItems, int64(1))
}

func (s *UnitTypeSuite) TestUnitTypeValidation() {
	res := s.MakeRequest("POST", "/api/v1/unit-types", map[string]interface{}{
		"property_id":    s.propertyID,
		"name":           "Negative UnitType",
		"code":           "NEG",
		"base_price":     -10.0,
		"max_occupancy":  2,
	}, s.token)
	s.Equal(http.StatusBadRequest, res.Code)

	res2 := s.MakeRequest("POST", "/api/v1/unit-types", map[string]interface{}{
		"property_id":    s.propertyID,
		"name":           "Zero Cap UnitType",
		"code":           "ZCP",
		"base_price":     100.0,
		"max_occupancy":  0,
	}, s.token)
	s.Equal(http.StatusBadRequest, res2.Code)

	res3 := s.MakeRequest("POST", "/api/v1/unit-types", map[string]interface{}{
		"property_id":    s.propertyID,
		"name":           "",
		"code":           "EMP",
		"base_price":     100.0,
		"max_occupancy":  2,
	}, s.token)
	s.Equal(http.StatusBadRequest, res3.Code)
}

func (s *UnitTypeSuite) TestUnitTypeNotFound() {
	res := s.MakeRequest("GET", "/api/v1/unit-types/00000000-0000-0000-0000-000000000000", nil, s.token)
	s.Equal(http.StatusNotFound, res.Code)

	resUpd := s.MakeRequest("PUT", "/api/v1/unit-types/00000000-0000-0000-0000-000000000000", map[string]interface{}{
		"base_price": 200.0,
	}, s.token)
	s.Equal(http.StatusNotFound, resUpd.Code)
}

func (s *UnitTypeSuite) TestUnitTypeDuplicateCode() {
	res := s.MakeRequest("POST", "/api/v1/unit-types", map[string]interface{}{
		"property_id":    s.propertyID,
		"name":           "Unique UnitType",
		"code":           "UNI",
		"base_price":     100.0,
		"total_quantity": 5,
		"max_occupancy":  2,
		"max_adults":     2,
		"max_children":   0,
		"amenities":      []string{"wifi"},
	}, s.token)
	s.Equal(http.StatusCreated, res.Code)

	res2 := s.MakeRequest("POST", "/api/v1/unit-types", map[string]interface{}{
		"property_id":    s.propertyID,
		"name":           "Another UnitType",
		"code":           "UNI",
		"base_price":     120.0,
		"total_quantity": 5,
		"max_occupancy":  2,
		"max_adults":     2,
		"max_children":   0,
		"amenities":      []string{"wifi"},
	}, s.token)
	s.Equal(http.StatusConflict, res2.Code)
}

func TestUnitTypeSuite(t *testing.T) {
	suite.Run(t, new(UnitTypeSuite))
}
