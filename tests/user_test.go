package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type UserSuite struct {
	BaseSuite
	ownerToken string
	orgID      string
}

func (s *UserSuite) SetupTest() {
	s.BaseSuite.SetupTest()
	s.ownerToken, s.orgID = s.GetAdminTokenAndOrg()
}

func (s *UserSuite) TestUserHierarchy() {
	res := s.MakeRequest("POST", "/api/v1/users", map[string]string{
		"organization_id": s.orgID,
		"email":           "manager@corp.com",
		"password":        "secret123",
		"role":            "manager",
		"first_name":      "Manager",
    "last_name":       "One",
    "phone":           "987654",
	}, s.ownerToken)
	s.Equal(http.StatusCreated, res.Code)

	resFail := s.MakeRequest("POST", "/api/v1/users", map[string]string{
		"organization_id": s.orgID,
		"email":           "another_owner@corp.com",
		"password":        "secret123",
		"role":            "owner",
		"first_name":      "Owner",
    "last_name":       "Two",
    "phone":           "987654",
	}, s.ownerToken)
	
	s.Equal(http.StatusForbidden, resFail.Code)
	s.Contains(resFail.Body.String(), "only super_admin can create organization owners")
}

func (s *UserSuite) TestListUsers() {
	res := s.MakeRequest("POST", "/api/v1/users", map[string]string{
		"organization_id": s.orgID,
		"email":           "staff@corp.com",
		"password":        "secret123",
		"role":            "staff",
		"first_name":      "Staff",
		"last_name":       "Member",
		"phone":           "123456",
	}, s.ownerToken)
	s.Equal(http.StatusCreated, res.Code)

	resList := s.MakeRequest("GET", "/api/v1/users?organization_id="+s.orgID, nil, s.ownerToken)
	s.Equal(http.StatusOK, resList.Code)

	var response entity.PaginatedResponse[entity.User]
	json.Unmarshal(resList.Body.Bytes(), &response)
	s.NotEmpty(response.Data)
	s.GreaterOrEqual(response.Meta.TotalItems, int64(1))
}

func (s *UserSuite) TestUserValidation() {
	res := s.MakeRequest("POST", "/api/v1/users", map[string]string{
		"organization_id": s.orgID,
		"email":           "invalid-email",
		"password":        "secret123",
		"role":            "staff",
		"first_name":      "Staff", "last_name": "Member", "phone": "123",
	}, s.ownerToken)
	s.Equal(http.StatusBadRequest, res.Code)

	res2 := s.MakeRequest("POST", "/api/v1/users", map[string]string{
		"organization_id": s.orgID,
		"email":           "valid@email.com",
		"password":        "",
		"role":            "staff",
		"first_name":      "Staff", "last_name": "Member", "phone": "123",
	}, s.ownerToken)
	s.Equal(http.StatusBadRequest, res2.Code)
}

func (s *UserSuite) TestUserNotFound() {
	res := s.MakeRequest("GET", "/api/v1/users/00000000-0000-0000-0000-000000000000", nil, s.ownerToken)
	s.Equal(http.StatusNotFound, res.Code)

	resUpd := s.MakeRequest("PUT", "/api/v1/users/00000000-0000-0000-0000-000000000000?organization_id="+s.orgID, map[string]string{
		"first_name": "Ghost",
	}, s.ownerToken)
	s.Equal(http.StatusNotFound, resUpd.Code)

	resDel := s.MakeRequest("DELETE", "/api/v1/users/00000000-0000-0000-0000-000000000000", nil, s.ownerToken)
	s.Equal(http.StatusNotFound, resDel.Code)
}

func (s *UserSuite) TestUserDuplicateEmail() {
	res := s.MakeRequest("POST", "/api/v1/users", map[string]string{
		"organization_id": s.orgID,
		"email":           "duplicate@corp.com",
		"password":        "pass1",
		"role":            "staff",
		"first_name":      "User", "last_name": "A", "phone": "1",
	}, s.ownerToken)
	s.Equal(http.StatusCreated, res.Code)

	res2 := s.MakeRequest("POST", "/api/v1/users", map[string]string{
		"organization_id": s.orgID,
		"email":           "duplicate@corp.com",
		"password":        "pass2",
		"role":            "staff",
		"first_name":      "User", "last_name": "B", "phone": "2",
	}, s.ownerToken)
	
	s.Equal(http.StatusConflict, res2.Code)
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserSuite))
}
