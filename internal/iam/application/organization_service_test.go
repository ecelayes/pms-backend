package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ecelayes/pms-backend/internal/iam/domain"
)

func TestOrganizationService_NewOrganizationService(t *testing.T) {
	svc := NewOrganizationService(nil)
	if svc == nil {
		t.Error("Expected non-nil service")
	}
}

func TestOrganizationService_Create_Success(t *testing.T) {
	repo := &mockOrgRepoForOrg{
		findByUserIDResult: nil,
	}
	svc := NewOrganizationService(repo)

	id, err := svc.Create(context.Background(), "Test Org", "TST")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id == "" {
		t.Error("Expected non-empty id")
	}
}

func TestOrganizationService_Create_InvalidOrg(t *testing.T) {
	repo := &mockOrgRepoForOrg{}
	svc := NewOrganizationService(repo)

	_, err := svc.Create(context.Background(), "", "")
	if err == nil {
		t.Error("Expected error for invalid org")
	}
}

func TestOrganizationService_GetByID_Success(t *testing.T) {
	org := domain.ReconstituteOrganization("org-1", "Test Org", "test.com", time.Now())
	repo := &mockOrgRepoForOrg{findByIDResult: org}
	svc := NewOrganizationService(repo)

	result, err := svc.GetByID(context.Background(), "org-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.ID() != "org-1" {
		t.Errorf("Expected org-1, got %s", result.ID())
	}
}

func TestOrganizationService_GetByID_NotFound(t *testing.T) {
	repo := &mockOrgRepoForOrg{findByIDResult: nil, findByIDErr: nil}
	svc := NewOrganizationService(repo)

	_, err := svc.GetByID(context.Background(), "non-existent")
	if err != ErrOrgNotFound {
		t.Errorf("Expected ErrOrgNotFound, got %v", err)
	}
}

func TestOrganizationService_GetAll_Success(t *testing.T) {
	orgs := []*domain.Organization{
		domain.ReconstituteOrganization("org-1", "Org 1", "org1.com", time.Now()),
		domain.ReconstituteOrganization("org-2", "Org 2", "org2.com", time.Now()),
	}
	repo := &mockOrgRepoForOrg{findAllResult: orgs}
	svc := NewOrganizationService(repo)

	result, err := svc.GetAll(context.Background())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 orgs, got %d", len(result))
	}
}

func TestOrganizationService_Update_Success(t *testing.T) {
	org := domain.ReconstituteOrganization("org-1", "Old Name", "old.com", time.Now())
	repo := &mockOrgRepoForOrg{findByIDResult: org}
	svc := NewOrganizationService(repo)

	err := svc.Update(context.Background(), "org-1", "New Name", "new.com")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestOrganizationService_Update_NotFound(t *testing.T) {
	repo := &mockOrgRepoForOrg{findByIDResult: nil}
	svc := NewOrganizationService(repo)

	err := svc.Update(context.Background(), "non-existent", "Name", "code")
	if err != ErrOrgNotFound {
		t.Errorf("Expected ErrOrgNotFound, got %v", err)
	}
}

func TestOrganizationService_Delete_Success(t *testing.T) {
	repo := &mockOrgRepoForOrg{}
	svc := NewOrganizationService(repo)

	err := svc.Delete(context.Background(), "org-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

type mockOrgRepoForOrg struct {
	findByUserIDResult *domain.Organization
	findByUserIDErr    error
	findByIDResult     *domain.Organization
	findByIDErr        error
	findAllResult      []*domain.Organization
	findAllErr         error
	deleteErr          error
	saveErr            error
}

func (m *mockOrgRepoForOrg) Save(ctx context.Context, o *domain.Organization) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	return nil
}
func (m *mockOrgRepoForOrg) FindByID(ctx context.Context, id string) (*domain.Organization, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	return m.findByIDResult, nil
}
func (m *mockOrgRepoForOrg) FindByUserID(ctx context.Context, userID string) (*domain.Organization, error) {
	if m.findByUserIDErr != nil {
		return nil, m.findByUserIDErr
	}
	return m.findByUserIDResult, nil
}
func (m *mockOrgRepoForOrg) FindByDomain(ctx context.Context, domain string) (*domain.Organization, error) {
	return nil, nil
}
func (m *mockOrgRepoForOrg) AddMember(ctx context.Context, orgID, userID, role string) error {
	return nil
}
func (m *mockOrgRepoForOrg) FindMemberRole(ctx context.Context, orgID, userID string) (string, error) {
	return "", nil
}
func (m *mockOrgRepoForOrg) FindAll(ctx context.Context) ([]*domain.Organization, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	return m.findAllResult, nil
}
func (m *mockOrgRepoForOrg) DeleteOrganization(ctx context.Context, id string) error {
	return m.deleteErr
}

func TestOrganizationService_Create_NewOrganizationError(t *testing.T) {
	repo := &mockOrgRepoForOrg{}
	svc := NewOrganizationService(repo)

	_, err := svc.Create(context.Background(), "", "")
	if err == nil {
		t.Error("Expected error from NewOrganization")
	}
}

func TestOrganizationService_Create_SaveError(t *testing.T) {
	repo := &mockOrgRepoForOrg{saveErr: errors.New("save error")}
	svc := NewOrganizationService(repo)

	_, err := svc.Create(context.Background(), "Test Org", "TST")
	if err == nil {
		t.Error("Expected error from save")
	}
}

func TestOrganizationService_GetByID_RepoError(t *testing.T) {
	repo := &mockOrgRepoForOrg{findByIDErr: errors.New("repo error")}
	svc := NewOrganizationService(repo)

	_, err := svc.GetByID(context.Background(), "org-1")
	if err == nil {
		t.Error("Expected error from repo")
	}
}
