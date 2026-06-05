package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// BcryptPasswordHasher is the production password hasher.
//
// SECURITY: Uses bcrypt with DefaultCost (currently 10).
// bcrypt is a slow, salted hash designed for password storage.
type BcryptPasswordHasher struct{}

func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{}
}

func (h *BcryptPasswordHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (h *BcryptPasswordHasher) Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// JWTTokenGenerator is the production token generator.
//
// SECURITY: Uses HS256 with user-supplied salt as the signing key.
// The salt comes from crypto/rand (32 bytes hex-encoded = 64 chars).
// Tokens have explicit expiration (24h auth, 15min reset).
// ParseUnsafe does NOT verify signature - callers must use VerifySignature.
type JWTTokenGenerator struct{}

func NewJWTTokenGenerator() *JWTTokenGenerator {
	return &JWTTokenGenerator{}
}

func (j *JWTTokenGenerator) GenerateAuthToken(userID, organizationID, role, userSalt string) (string, error) {
	claims := Claims{
		UserID:         userID,
		OrganizationID: organizationID,
		Role:           role,
		Purpose:        PurposeAuth,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(userSalt))
}

func (j *JWTTokenGenerator) GenerateResetToken(userID, userSalt string) (string, error) {
	claims := Claims{
		UserID:  userID,
		Role:    "none",
		Purpose: PurposeReset,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(userSalt))
}

func (j *JWTTokenGenerator) ParseUnsafe(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}
	return token.Claims.(*Claims), nil
}

func (j *JWTTokenGenerator) VerifySignature(tokenString, userSalt string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// SECURITY: Reject "none" algorithm - prevent alg confusion attacks
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(userSalt), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token signature")
	}
	return claims, nil
}

// CryptoRandomSaltGenerator is the production random generator.
//
// SECURITY: Uses crypto/rand (CSPRNG). The salt is 32 bytes = 256 bits of entropy.
type CryptoRandomSaltGenerator struct {
	reader io.Reader
}

func NewCryptoRandomSaltGenerator() *CryptoRandomSaltGenerator {
	return &CryptoRandomSaltGenerator{reader: rand.Reader}
}

func newCryptoRandomSaltGeneratorWithReader(r io.Reader) *CryptoRandomSaltGenerator {
	return &CryptoRandomSaltGenerator{reader: r}
}

func (g *CryptoRandomSaltGenerator) Generate() (string, error) {
	bytes := make([]byte, 32)
	if _, err := g.reader.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
