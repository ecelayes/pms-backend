package auth

import (
	"io"
	"strings"
	"testing"
	"time"
)

func TestBcryptPasswordHasher_HashAndVerify(t *testing.T) {
	h := NewBcryptPasswordHasher()
	password := "super-secret-password"
	hash, err := h.Hash(password)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}
	if hash == "" {
		t.Fatal("empty hash returned")
	}
	if !h.Verify(password, hash) {
		t.Error("Verify returned false for correct password")
	}
	if h.Verify("wrong-password", hash) {
		t.Error("Verify returned true for wrong password")
	}
}

func TestBcryptPasswordHasher_HashDifferentForSamePassword(t *testing.T) {
	h := NewBcryptPasswordHasher()
	password := "same-password"
	hash1, _ := h.Hash(password)
	hash2, _ := h.Hash(password)
	if hash1 == hash2 {
		t.Error("bcrypt should produce different hashes for same password (due to salt)")
	}
}

func TestBcryptPasswordHasher_VerifyInvalidHash(t *testing.T) {
	h := NewBcryptPasswordHasher()
	if h.Verify("password", "not-a-valid-hash") {
		t.Error("Verify returned true for invalid hash")
	}
}

func TestJWTTokenGenerator_GenerateAuthToken_Success(t *testing.T) {
	g := NewJWTTokenGenerator()
	salt := "test-salt-1234567890"
	token, err := g.GenerateAuthToken("u-1", "o-1", "admin", salt)
	if err != nil {
		t.Fatalf("GenerateAuthToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("empty token")
	}
	if !strings.Contains(token, ".") {
		t.Error("token missing JWT separator")
	}
}

func TestJWTTokenGenerator_GenerateResetToken_Success(t *testing.T) {
	g := NewJWTTokenGenerator()
	salt := "test-salt-1234567890"
	token, err := g.GenerateResetToken("u-1", salt)
	if err != nil {
		t.Fatalf("GenerateResetToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("empty token")
	}
}

func TestJWTTokenGenerator_VerifySignature_Valid(t *testing.T) {
	g := NewJWTTokenGenerator()
	salt := "test-salt-1234567890"
	token, err := g.GenerateAuthToken("u-1", "o-1", "admin", salt)
	if err != nil {
		t.Fatalf("GenerateAuthToken failed: %v", err)
	}
	claims, err := g.VerifySignature(token, salt)
	if err != nil {
		t.Fatalf("VerifySignature failed: %v", err)
	}
	if claims.UserID != "u-1" {
		t.Errorf("expected UserID u-1, got %s", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Errorf("expected role admin, got %s", claims.Role)
	}
	if claims.Purpose != PurposeAuth {
		t.Errorf("expected purpose auth, got %s", claims.Purpose)
	}
}

func TestJWTTokenGenerator_VerifySignature_WrongSalt(t *testing.T) {
	g := NewJWTTokenGenerator()
	salt := "test-salt-1234567890"
	token, _ := g.GenerateAuthToken("u-1", "o-1", "admin", salt)
	_, err := g.VerifySignature(token, "wrong-salt-1234567890")
	if err == nil {
		t.Error("VerifySignature should fail with wrong salt")
	}
}

func TestJWTTokenGenerator_VerifySignature_TamperedToken(t *testing.T) {
	g := NewJWTTokenGenerator()
	salt := "test-salt-1234567890"
	token, _ := g.GenerateAuthToken("u-1", "o-1", "admin", salt)
	tampered := token + "tamper"
	_, err := g.VerifySignature(tampered, salt)
	if err == nil {
		t.Error("VerifySignature should fail with tampered token")
	}
}

func TestJWTTokenGenerator_VerifySignature_NoneAlgAttack(t *testing.T) {
	// SECURITY: Verify rejection of "alg: none" attack
	// Token is built manually with alg: none to test our protection
	noneToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJVc2VySUQiOiJ1LTEiLCJQdXJwb3NlIjoiYXV0aCJ9."
	g := NewJWTTokenGenerator()
	_, err := g.VerifySignature(noneToken, "any-salt")
	if err == nil {
		t.Error("VerifySignature must reject 'none' algorithm")
	}
}

func TestJWTTokenGenerator_ParseUnsafe_Success(t *testing.T) {
	g := NewJWTTokenGenerator()
	salt := "test-salt-1234567890"
	token, _ := g.GenerateAuthToken("u-1", "o-1", "admin", salt)
	claims, err := g.ParseUnsafe(token)
	if err != nil {
		t.Fatalf("ParseUnsafe failed: %v", err)
	}
	if claims.UserID != "u-1" {
		t.Errorf("expected UserID u-1, got %s", claims.UserID)
	}
}

func TestJWTTokenGenerator_ParseUnsafe_Invalid(t *testing.T) {
	g := NewJWTTokenGenerator()
	_, err := g.ParseUnsafe("not.a.token")
	if err == nil {
		t.Error("ParseUnsafe should fail on invalid token")
	}
}

func TestJWTTokenGenerator_AuthTokenExpiry(t *testing.T) {
	g := NewJWTTokenGenerator()
	salt := "test-salt-1234567890"
	token, _ := g.GenerateAuthToken("u-1", "o-1", "admin", salt)
	claims, _ := g.ParseUnsafe(token)
	if claims.ExpiresAt == nil {
		t.Fatal("auth token missing expiry")
	}
	expectedExpiry := time.Now().Add(24 * time.Hour)
	diff := claims.ExpiresAt.Time.Sub(expectedExpiry)
	if diff < -1*time.Minute || diff > 1*time.Minute {
		t.Errorf("auth token expiry out of expected range: %v", diff)
	}
}

func TestJWTTokenGenerator_ResetTokenExpiry(t *testing.T) {
	g := NewJWTTokenGenerator()
	salt := "test-salt-1234567890"
	token, _ := g.GenerateResetToken("u-1", salt)
	claims, _ := g.ParseUnsafe(token)
	if claims.ExpiresAt == nil {
		t.Fatal("reset token missing expiry")
	}
	expectedExpiry := time.Now().Add(15 * time.Minute)
	diff := claims.ExpiresAt.Time.Sub(expectedExpiry)
	if diff < -1*time.Minute || diff > 1*time.Minute {
		t.Errorf("reset token expiry out of expected range: %v", diff)
	}
}

func TestJWTTokenGenerator_ResetTokenHasNoOrg(t *testing.T) {
	g := NewJWTTokenGenerator()
	salt := "test-salt-1234567890"
	token, _ := g.GenerateResetToken("u-1", salt)
	claims, _ := g.ParseUnsafe(token)
	if claims.OrganizationID != "" {
		t.Errorf("reset token should not have org, got %s", claims.OrganizationID)
	}
}

func TestCryptoRandomSaltGenerator_Generate(t *testing.T) {
	g := NewCryptoRandomSaltGenerator()
	salt, err := g.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	// 32 bytes hex-encoded = 64 chars
	if len(salt) != 64 {
		t.Errorf("expected 64-char salt, got %d", len(salt))
	}
}

func TestCryptoRandomSaltGenerator_DifferentOutputs(t *testing.T) {
	g := NewCryptoRandomSaltGenerator()
	salt1, _ := g.Generate()
	salt2, _ := g.Generate()
	if salt1 == salt2 {
		t.Error("CSPRNG should produce different salts")
	}
}

func TestClaims_PurposeConstants(t *testing.T) {
	if PurposeAuth != "auth" {
		t.Errorf("PurposeAuth constant mismatch: %s", PurposeAuth)
	}
	if PurposeReset != "reset" {
		t.Errorf("PurposeReset constant mismatch: %s", PurposeReset)
	}
}

// TestBcryptPasswordHasher_Hash_PasswordTooLong verifies that bcrypt's
// 72-byte input limit surfaces as an error. Callers must reject or
// pre-hash long passwords before reaching this layer.
func TestBcryptPasswordHasher_Hash_PasswordTooLong(t *testing.T) {
	h := NewBcryptPasswordHasher()
	// 73 bytes (1 over the bcrypt limit)
	password := strings.Repeat("a", 73)
	_, err := h.Hash(password)
	if err == nil {
		t.Error("expected error for password > 72 bytes, got nil")
	}
}

// TestCryptoRandomSaltGenerator_Generate_ReaderError verifies the error
// branch when the underlying reader fails. Uses an injected reader because
// crypto/rand cannot be made to fail in production. Production wiring
// (NewCryptoRandomSaltGenerator) is unaffected.
type errReader struct{}

func (errReader) Read(_ []byte) (int, error) { return 0, io.EOF }

func TestCryptoRandomSaltGenerator_Generate_ReaderError(t *testing.T) {
	g := newCryptoRandomSaltGeneratorWithReader(errReader{})
	salt, err := g.Generate()
	if err == nil {
		t.Fatal("expected error from failing reader, got nil")
	}
	if salt != "" {
		t.Errorf("expected empty salt on error, got %q", salt)
	}
}
