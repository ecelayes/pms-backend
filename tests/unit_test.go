package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/stretchr/testify/suite"
)

type UnitSuite struct {
	BaseSuite
	token      string
	orgID      string
	propertyID string
	unitTypeID string
}

func TestUnitSuite(t *testing.T) {
	suite.Run(t, new(UnitSuite))
}

func (s *UnitSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.token, s.orgID = s.BaseSuite.GetAdminTokenAndOrg()

	propReq := entity.CreatePropertyRequest{
		OrganizationID: s.orgID,
		Name:    "Unit Test Property",
		Code:    "UTP01",
		Type:    "HOTEL",
	}
	resp := s.MakeRequest("POST", "/api/v1/properties", propReq, s.token)
	s.Require().Equal(http.StatusCreated, resp.Code)
	
	var propResult map[string]string
	json.Unmarshal(resp.Body.Bytes(), &propResult)
	s.propertyID = propResult["property_id"]

	utReq := entity.CreateUnitTypeRequest{
		PropertyID:    s.propertyID,
		Name:          "Standard Room",
		Code:          "STD",
		BasePrice:     100.0,
		MaxAdults:     2,
		MaxChildren:   1,
		MaxOccupancy:  3,
		TotalQuantity: 10,
	}
	resp = s.MakeRequest("POST", "/api/v1/unit-types", utReq, s.token)
	s.Require().Equal(http.StatusCreated, resp.Code)
	
	var utResult map[string]string
	json.Unmarshal(resp.Body.Bytes(), &utResult)
	s.unitTypeID = utResult["unit_type_id"]
}

func (s *UnitSuite) TestCreateUnit() {
	req := entity.CreateUnitRequest{
		PropertyID: s.propertyID,
		UnitTypeID: s.unitTypeID,
		Name:       "101",
		Status:     "CLEAN",
	}

	resp := s.MakeRequest("POST", "/api/v1/units", req, s.token)
	s.Equal(http.StatusCreated, resp.Code)
	
	var result map[string]string
	json.Unmarshal(resp.Body.Bytes(), &result)
	s.NotEmpty(result["unit_id"])
}

func (s *UnitSuite) TestGetUnitsByProperty() {
	req := entity.CreateUnitRequest{
		PropertyID: s.propertyID,
		UnitTypeID: s.unitTypeID,
		Name:       "102",
		Status:     "CLEAN",
	}
	s.MakeRequest("POST", "/api/v1/units", req, s.token)

	resp := s.MakeRequest("GET", "/api/v1/units?property_id=" + s.propertyID, nil, s.token)
	s.Equal(http.StatusOK, resp.Code)

	var units []entity.Unit
	json.Unmarshal(resp.Body.Bytes(), &units)
	s.GreaterOrEqual(len(units), 1)
	found := false
	for _, u := range units {
		if u.Name == "102" {
			found = true
			break
		}
	}
	s.True(found)
}

func (s *UnitSuite) TestUpdateUnit() {
	req := entity.CreateUnitRequest{
		PropertyID: s.propertyID,
		UnitTypeID: s.unitTypeID,
		Name:       "103",
		Status:     "CLEAN",
	}
	createResp := s.MakeRequest("POST", "/api/v1/units", req, s.token)
	s.Require().Equal(http.StatusCreated, createResp.Code)
	
	var createResult map[string]string
	json.Unmarshal(createResp.Body.Bytes(), &createResult)
	unitID := createResult["unit_id"]

	updateReq := entity.UpdateUnitRequest{
		Name:   "103-B",
		Status: "DIRTY",
	}
	updateResp := s.MakeRequest("PUT", "/api/v1/units/" + unitID, updateReq, s.token)
	s.Equal(http.StatusOK, updateResp.Code)

	getResp := s.MakeRequest("GET", "/api/v1/units?property_id=" + s.propertyID, nil, s.token)
	var units []entity.Unit
	json.Unmarshal(getResp.Body.Bytes(), &units)
	
	var updatedUnit entity.Unit
	for _, u := range units {
		if u.ID == unitID {
			updatedUnit = u
			break
		}
	}
	s.Equal("103-B", updatedUnit.Name)
	s.Equal("DIRTY", updatedUnit.Status)
}

func (s *UnitSuite) TestDeleteUnit() {
	req := entity.CreateUnitRequest{
		PropertyID: s.propertyID,
		UnitTypeID: s.unitTypeID,
		Name:       "104",
		Status:     "CLEAN",
	}
	createResp := s.MakeRequest("POST", "/api/v1/units", req, s.token)
	s.Require().Equal(http.StatusCreated, createResp.Code)
	
	var createResult map[string]string
	json.Unmarshal(createResp.Body.Bytes(), &createResult)
	unitID := createResult["unit_id"]

	delResp := s.MakeRequest("DELETE", "/api/v1/units/" + unitID, nil, s.token)
	s.Equal(http.StatusOK, delResp.Code)

	getResp := s.MakeRequest("GET", "/api/v1/units?property_id=" + s.propertyID, nil, s.token)
	var units []entity.Unit
	json.Unmarshal(getResp.Body.Bytes(), &units)
	
	for _, u := range units {
		s.NotEqual(unitID, u.ID)
	}
}
