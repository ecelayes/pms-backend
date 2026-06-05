package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ecelayes/pms-backend/internal/iam/domain"
)

type mockUserRepoForUserService struct {
	users                map[string]*domain.User
	saveErr              error
	findByIDResult       *domain.User
	findByIDErr          error
	findByEmailResult    *domain.User
	findByEmailErr       error
	findByEmailCallCount int
	findByEmailFunc      func(ctx context.Context, email string) (*domain.User, error)
	findAllResult        []*domain.User
	findAllErr           error
	deleteErr            error
	ensureGuestErr       error
}

func (m *mockUserRepoForUserService) Save(ctx context.Context, u *domain.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	if m.users == nil {
		m.users = make(map[string]*domain.User)
	}
	m.users[u.ID()] = u
	return nil
}
func (m *mockUserRepoForUserService) FindByID(ctx context.Context, id string) (*domain.User, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	if m.findByIDResult != nil {
		return m.findByIDResult, nil
	}
	if m.users != nil {
		return m.users[id], nil
	}
	return nil, nil
}
func (m *mockUserRepoForUserService) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	m.findByEmailCallCount++
	if m.findByEmailFunc != nil {
		return m.findByEmailFunc(ctx, email)
	}
	if m.findByEmailErr != nil {
		return nil, m.findByEmailErr
	}
	return m.findByEmailResult, nil
}
func (m *mockUserRepoForUserService) FindAllByOrganization(ctx context.Context, orgID string) ([]*domain.User, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	return m.findAllResult, nil
}
func (m *mockUserRepoForUserService) DeleteUser(ctx context.Context, id string) error {
	return m.deleteErr
}
func (m *mockUserRepoForUserService) EnsureGuest(ctx context.Context, u *domain.User) error {
	return m.ensureGuestErr
}

type mockOrgRepoForUserService struct {
	addMemberErr    error
	}

func (m *mockOrgRepoForUserService) Save(ctx context.Context, o *domain.Organization) error {
	return nil
}
func (m *mockOrgRepoForUserService) FindByID(ctx context.Context, id string) (*domain.Organization, error) {
	return nil, nil
}
func (m *mockOrgRepoForUserService) FindByUserID(ctx context.Context, userID string) (*domain.Organization, error) {
	return nil, nil
}
func (m *mockOrgRepoForUserService) FindByDomain(ctx context.Context, domain string) (*domain.Organization, error) {
	return nil, nil
}
func (m *mockOrgRepoForUserService) AddMember(ctx context.Context, orgID, userID, role string) error {
	return m.addMemberErr
}
func (m *mockOrgRepoForUserService) FindMemberRole(ctx context.Context, orgID, userID string) (string, error) {
	return "", nil
}
func (m *mockOrgRepoForUserService) FindAll(ctx context.Context) ([]*domain.Organization, error) {
	return nil, nil
}
func (m *mockOrgRepoForUserService) DeleteOrganization(ctx context.Context, id string) error {
	return nil
}

func TestUserService_Register_Success(t *testing.T) {
	repo := &mockUserRepoForUserService{}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	id, err := svc.Register(context.Background(), "org-1", "test@test.com", "Good.Pass1", "admin", "John", "Doe", "123")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id == "" {
		t.Error("Expected non-empty id")
	}
}

func TestUserService_Register_SuperAdmin(t *testing.T) {
	repo := &mockUserRepoForUserService{}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	id, err := svc.Register(context.Background(), "", "super@test.com", "Good.Pass1", "super_admin", "Super", "Admin", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id == "" {
		t.Error("Expected non-empty id")
	}
}

func TestUserService_Register_SaveError(t *testing.T) {
	repo := &mockUserRepoForUserService{saveErr: errors.New("save error")}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	_, err := svc.Register(context.Background(), "", "test@test.com", "Good.Pass1", "user", "John", "Doe", "")
	if err == nil {
		t.Error("Expected error from save")
	}
}

func TestUserService_Register_WithOrg(t *testing.T) {
	repo := &mockUserRepoForUserService{}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	id, err := svc.Register(context.Background(), "org-123", "test@test.com", "Good.Pass1", "admin", "John", "Doe", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id == "" {
		t.Error("Expected non-empty id")
	}
}

func TestUserService_Register_RejectsWeakPassword(t *testing.T) {
	repo := &mockUserRepoForUserService{}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	for _, p := range []string{"", "short", "password123", "12345678", "alllowercase"} {
		_, err := svc.Register(context.Background(), "org-123", "test@test.com", p, "user", "John", "Doe", "")
		if err == nil {
			t.Errorf("expected ErrWeakPassword for %q, got nil", p)
		}
	}
}

func TestUserService_GetByID_Success(t *testing.T) {
	user := domain.ReconstituteUser("user-1", "test@test.com", "hash", "salt", "admin", "John", "Doe", "", time.Now())
	repo := &mockUserRepoForUserService{findByIDResult: user}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	result, err := svc.GetByID(context.Background(), "user-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.ID() != "user-1" {
		t.Errorf("Expected user-1, got %s", result.ID())
	}
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	repo := &mockUserRepoForUserService{findByIDResult: nil}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	_, err := svc.GetByID(context.Background(), "non-existent")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_GetByID_RepoError(t *testing.T) {
	repo := &mockUserRepoForUserService{findByIDErr: errors.New("repo error")}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	_, err := svc.GetByID(context.Background(), "user-1")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestUserService_GetByID_NotFoundError(t *testing.T) {
	repo := &mockUserRepoForUserService{findByIDErr: domain.ErrNotFound}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	_, err := svc.GetByID(context.Background(), "user-1")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_GetAll_Success(t *testing.T) {
	users := []*domain.User{
		domain.ReconstituteUser("user-1", "test1@test.com", "hash", "salt", "admin", "John", "Doe", "", time.Now()),
		domain.ReconstituteUser("user-2", "test2@test.com", "hash", "salt", "user", "Jane", "Doe", "", time.Now()),
	}
	repo := &mockUserRepoForUserService{findAllResult: users}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	result, err := svc.GetAll(context.Background(), "org-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 users, got %d", len(result))
	}
}

func TestUserService_GetAll_RepoError(t *testing.T) {
	repo := &mockUserRepoForUserService{findAllErr: errors.New("repo error")}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	_, err := svc.GetAll(context.Background(), "org-1")
	if err == nil {
		t.Error("Expected error from repo")
	}
}

func TestUserService_Update_Success(t *testing.T) {
	user := domain.ReconstituteUser("user-1", "test@test.com", "hash", "salt", "user", "John", "Doe", "", time.Now())
	repo := &mockUserRepoForUserService{users: map[string]*domain.User{"user-1": user}}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	err := svc.Update(context.Background(), "user-1", "admin", "Jane", "Smith", "456")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestUserService_Update_NotFound(t *testing.T) {
	repo := &mockUserRepoForUserService{}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	err := svc.Update(context.Background(), "non-existent", "admin", "Jane", "Smith", "456")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_Delete_Success(t *testing.T) {
	user := domain.ReconstituteUser("user-1", "test@test.com", "hash", "salt", "admin", "John", "Doe", "", time.Now())
	repo := &mockUserRepoForUserService{users: map[string]*domain.User{"user-1": user}}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	err := svc.Delete(context.Background(), "user-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestUserService_Delete_NotFound(t *testing.T) {
	repo := &mockUserRepoForUserService{}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	err := svc.Delete(context.Background(), "non-existent")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_FindOrCreateGuest_ExistingUser(t *testing.T) {
	user := domain.ReconstituteUser("user-1", "test@test.com", "hash", "salt", "guest", "John", "Doe", "", time.Now())
	repo := &mockUserRepoForUserService{findByEmailResult: user}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	id, err := svc.FindOrCreateGuest(context.Background(), "test@test.com", "John", "Doe", "123")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id != "user-1" {
		t.Errorf("Expected user-1, got %s", id)
	}
}

func TestUserService_FindOrCreateGuest_NewUser(t *testing.T) {
	repo := &mockUserRepoForUserService{findByEmailResult: nil}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	id, err := svc.FindOrCreateGuest(context.Background(), "new@test.com", "New", "User", "123")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id == "" {
		t.Error("Expected non-empty id")
	}
}

func TestUserService_FindOrCreateGuest_EmailError(t *testing.T) {
	repo := &mockUserRepoForUserService{findByEmailErr: errors.New("email error")}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	_, err := svc.FindOrCreateGuest(context.Background(), "test@test.com", "John", "Doe", "123")
	if err == nil {
		t.Error("Expected error from repo")
	}
}


func TestUserService_FindOrCreateGuest_EnsureGuestError(t *testing.T) {
	user := domain.ReconstituteUser("user-1", "test@test.com", "hash", "salt", "guest", "John", "Doe", "", time.Now())
	repo := &mockUserRepoForUserService{findByEmailResult: user, ensureGuestErr: errors.New("ensure guest error")}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	_, err := svc.FindOrCreateGuest(context.Background(), "test@test.com", "John", "Doe", "123")
	if err == nil {
		t.Error("Expected error from EnsureGuest")
	}
}

func TestUserService_Register_NewUserError(t *testing.T) {
	repo := &mockUserRepoForUserService{}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	_, err := svc.Register(context.Background(), "org-1", "not-an-email", "Good.Pass1", "user", "John", "Doe", "")
	if err == nil {
		t.Error("Expected error from NewUser")
	}
}

func TestUserService_Register_AddMemberError(t *testing.T) {
	repo := &mockUserRepoForUserService{}
	orgRepo := &mockOrgRepoForUserService{addMemberErr: errors.New("add member error")}
	svc := newUserServiceForTest(repo, orgRepo)

	_, err := svc.Register(context.Background(), "org-1", "test@test.com", "Good.Pass1", "admin", "John", "Doe", "123")
	if err == nil {
		t.Error("Expected error from AddMember")
	}
}

func TestUserService_Update_SaveError(t *testing.T) {
	user := domain.ReconstituteUser("user-1", "test@test.com", "hash", "salt", "user", "John", "Doe", "", time.Now())
	repo := &mockUserRepoForUserService{findByIDResult: user, saveErr: errors.New("save error")}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	err := svc.Update(context.Background(), "user-1", "admin", "Jane", "Doe", "456")
	if err == nil {
		t.Error("Expected error from Save")
	}
}

func TestUserService_Delete_RepoError(t *testing.T) {
	repo := &mockUserRepoForUserService{deleteErr: errors.New("delete error")}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	err := svc.Delete(context.Background(), "user-1")
	if err == nil {
		t.Error("Expected error from Delete")
	}
}

func TestUserService_FindOrCreateGuest_RegisterError(t *testing.T) {
	// Test that any non-duplicate error from Register is propagated
	repo := &mockUserRepoForUserService{findByEmailResult: nil, saveErr: errors.New("some other error")}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	_, err := svc.FindOrCreateGuest(context.Background(), "dup@test.com", "John", "Doe", "123")
	if err == nil {
		t.Error("Expected error from Register")
	}
	if err.Error() != "some other error" {
		t.Errorf("Expected 'some other error', got: %v", err)
	}
}

func TestUserService_FindOrCreateGuest_DuplicateRetrySuccess(t *testing.T) {
	// First FindByEmail returns nil (no user), then Register fails with "duplicate",
	// then FindByEmail returns the user
	repo := &mockUserRepoForUserService{
		findByEmailResult: nil,
		findByEmailCallCount: 0,
		saveErr:            errors.New("duplicate key violation"),
	}
	// Track that FindByEmail is called multiple times
	repo.findByEmailFunc = func(ctx context.Context, email string) (*domain.User, error) {
		if repo.findByEmailCallCount == 1 {
			return nil, nil
		}
		return domain.ReconstituteUser("user-dup", email, "hash", "salt", "guest", "John", "Doe", "", time.Now()), nil
	}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	id, err := svc.FindOrCreateGuest(context.Background(), "dup@test.com", "John", "Doe", "123")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id == "" {
		t.Error("Expected non-empty id")
	}
}

func TestUserService_FindOrCreateGuest_DuplicateRetryFindError(t *testing.T) {
	repo := &mockUserRepoForUserService{
		findByEmailResult: nil,
		saveErr:            errors.New("duplicate key violation"),
		findByEmailErr:     errors.New("find error on retry"),
	}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	_, err := svc.FindOrCreateGuest(context.Background(), "dup@test.com", "John", "Doe", "123")
	if err == nil {
		t.Error("Expected error from retry find")
	}
}

func TestUserService_FindOrCreateGuest_NewUserFindByIDError(t *testing.T) {
	repo := &mockUserRepoForUserService{findByEmailResult: nil, findByIDErr: errors.New("find by id error")}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	_, err := svc.FindOrCreateGuest(context.Background(), "new@test.com", "John", "Doe", "123")
	if err == nil {
		t.Error("Expected error from FindByID")
	}
}

func TestUserService_FindOrCreateGuest_NewUserEnsureGuestError(t *testing.T) {
	repo := &mockUserRepoForUserService{findByEmailResult: nil, ensureGuestErr: errors.New("ensure guest error on new user")}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	_, err := svc.FindOrCreateGuest(context.Background(), "new@test.com", "John", "Doe", "123")
	if err == nil {
		t.Error("Expected error from EnsureGuest")
	}
}

func TestUserService_Register_SaltError(t *testing.T) {
	repo := &mockUserRepoForUserService{}
	orgRepo := &mockOrgRepoForUserService{}
	hasher := &mockPasswordHasher{hashResult: "hashed"}
	sg := &mockSaltGenerator{generateErr: errors.New("salt error")}
	svc := newUserServiceWithMocks(repo, orgRepo, hasher, sg)

	_, err := svc.Register(context.Background(), "", "test@test.com", "Good.Pass1", "user", "John", "Doe", "123")
	if err == nil {
		t.Error("Expected salt error")
	}
}

func TestUserService_Register_HashError(t *testing.T) {
	repo := &mockUserRepoForUserService{}
	orgRepo := &mockOrgRepoForUserService{}
	hasher := &mockPasswordHasher{hashErr: errors.New("hash error")}
	sg := &mockSaltGenerator{generateRes: "salt"}
	svc := newUserServiceWithMocks(repo, orgRepo, hasher, sg)

	_, err := svc.Register(context.Background(), "", "test@test.com", "Good.Pass1", "user", "John", "Doe", "123")
	if err == nil {
		t.Error("Expected hash error")
	}
}

func TestUserService_FindOrCreateGuest_EnsureGuestErrorOnRetry(t *testing.T) {
	repo := &mockUserRepoForUserService{
		findByEmailCallCount: 0,
		saveErr:              errors.New("duplicate key violation"),
		ensureGuestErr:       errors.New("ensure guest error"),
	}
	repo.findByEmailFunc = func(ctx context.Context, email string) (*domain.User, error) {
		if repo.findByEmailCallCount == 1 {
			return nil, nil
		}
		return domain.ReconstituteUser("user-dup", email, "hash", "salt", "guest", "John", "Doe", "", time.Now()), nil
	}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	_, err := svc.FindOrCreateGuest(context.Background(), "dup@test.com", "John", "Doe", "123")
	if err == nil {
		t.Error("Expected ensure guest error")
	}
}

func TestUserService_FindOrCreateGuest_DuplicateRetryFindError2(t *testing.T) {
	repo := &mockUserRepoForUserService{
		findByEmailCallCount: 0,
		saveErr:              errors.New("duplicate key violation"),
	}
	repo.findByEmailFunc = func(ctx context.Context, email string) (*domain.User, error) {
		if repo.findByEmailCallCount == 1 {
			return nil, nil
		}
		return nil, errors.New("retry find error")
	}
	orgRepo := &mockOrgRepoForUserService{}
	svc := newUserServiceForTest(repo, orgRepo)

	_, err := svc.FindOrCreateGuest(context.Background(), "dup@test.com", "John", "Doe", "123")
	if err == nil {
		t.Error("Expected retry find error")
	}
}
