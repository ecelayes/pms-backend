package application

import (
	"github.com/ecelayes/pms-backend/internal/iam/domain"
	"github.com/ecelayes/pms-backend/pkg/auth"
)

// newAuthServiceForTest creates an AuthService with standard crypto mocks.
// Use this for tests that don't care about crypto behavior.
func newAuthServiceForTest(userRepo domain.UserRepository, orgRepo domain.OrganizationRepository, emailSvc EmailService) *AuthService {
	hasher, tg, sg := newStandardMocks()
	return NewAuthService(userRepo, orgRepo, emailSvc, hasher, tg, sg)
}

// newUserServiceForTest creates a UserService with standard crypto mocks.
func newUserServiceForTest(repo domain.UserRepository, orgRepo domain.OrganizationRepository) *UserService {
	return newUserServiceWithMocks(repo, orgRepo, &mockPasswordHasher{hashResult: "hashed"}, &mockSaltGenerator{generateRes: "test-salt"})
}

// newUserServiceWithMocks creates a UserService with custom crypto mocks.
func newUserServiceWithMocks(repo domain.UserRepository, orgRepo domain.OrganizationRepository, hasher auth.PasswordHasher, sg auth.RandomSaltGenerator) *UserService {
	return NewUserService(repo, orgRepo, hasher, sg)
}

// newStandardMocks returns the standard mock set used by newAuthServiceForTest.
func newStandardMocks() (auth.PasswordHasher, auth.TokenGenerator, auth.RandomSaltGenerator) {
	return &mockPasswordHasher{hashResult: "hashed", verifyResult: true},
		&mockTokenGenerator{
			generateAuthResult:  "test-token",
			generateResetResult: "test-reset-token",
			parseUnsafeResult:  &auth.Claims{},
			verifySigResult:    &auth.Claims{},
		},
		&mockSaltGenerator{generateRes: "test-salt"}
}
