package application

import (
	"github.com/ecelayes/pms-backend/pkg/auth"
)

// Test mocks for the crypto interfaces.
// These allow us to simulate error paths in the application services
// without touching the real bcrypt/JWT/crypto-rand implementations.

// mockPasswordHasher - configurable mock for PasswordHasher
type mockPasswordHasher struct {
	hashErr      error
	hashResult   string
	hashCalls    int
	verifyResult bool
	}

func (m *mockPasswordHasher) Hash(password string) (string, error) {
	m.hashCalls++
	if m.hashErr != nil {
		return "", m.hashErr
	}
	return m.hashResult, nil
}

func (m *mockPasswordHasher) Verify(password, hash string) bool {
	return m.verifyResult
}

// mockTokenGenerator - configurable mock for TokenGenerator
type mockTokenGenerator struct {
	generateAuthErr    error
	generateAuthResult string
	generateResetErr   error
	generateResetResult string
	parseUnsafeErr     error
	parseUnsafeResult  *auth.Claims
	verifySigErr       error
	verifySigResult    *auth.Claims
}

func (m *mockTokenGenerator) GenerateAuthToken(userID, organizationID, role, userSalt string) (string, error) {
	if m.generateAuthErr != nil {
		return "", m.generateAuthErr
	}
	return m.generateAuthResult, nil
}

func (m *mockTokenGenerator) GenerateResetToken(userID, userSalt string) (string, error) {
	if m.generateResetErr != nil {
		return "", m.generateResetErr
	}
	return m.generateResetResult, nil
}

func (m *mockTokenGenerator) ParseUnsafe(tokenString string) (*auth.Claims, error) {
	if m.parseUnsafeErr != nil {
		return nil, m.parseUnsafeErr
	}
	return m.parseUnsafeResult, nil
}

func (m *mockTokenGenerator) VerifySignature(tokenString, userSalt string) (*auth.Claims, error) {
	if m.verifySigErr != nil {
		return nil, m.verifySigErr
	}
	return m.verifySigResult, nil
}

// mockSaltGenerator - configurable mock for RandomSaltGenerator
type mockSaltGenerator struct {
	generateErr error
	generateRes string
	generateCnt int
}

func (m *mockSaltGenerator) Generate() (string, error) {
	m.generateCnt++
	if m.generateErr != nil {
		return "", m.generateErr
	}
	if m.generateRes != "" {
		return m.generateRes, nil
	}
	return "mock-salt", nil
}
