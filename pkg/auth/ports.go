package auth

// PasswordHasher is the contract for password hashing implementations.
//
// SECURITY: Implementations MUST use a slow, salted hash function
// (e.g., bcrypt, scrypt, argon2id). Plain SHA/MD5/NOT-SALTED implementations
// are unsafe and will be rejected by code review.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) bool
}

// TokenGenerator is the contract for JWT token generation/validation.
//
// SECURITY: Implementations MUST:
//   - Sign tokens with a strong algorithm (HS256 minimum, RS256/ES256 preferred)
//   - Validate the signature on every Verify call (never trust unsigned claims)
//   - Set explicit expiration times
//   - Reject tokens with "alg: none"
type TokenGenerator interface {
	GenerateAuthToken(userID, organizationID, role, userSalt string) (string, error)
	GenerateResetToken(userID, userSalt string) (string, error)
	ParseUnsafe(tokenString string) (*Claims, error)
	VerifySignature(tokenString, userSalt string) (*Claims, error)
}

// RandomSaltGenerator is the contract for cryptographically secure random generation.
//
// SECURITY: Implementations MUST use crypto/rand (or equivalent CSPRNG).
// math/rand is NOT acceptable for security-sensitive contexts.
type RandomSaltGenerator interface {
	Generate() (string, error)
}
