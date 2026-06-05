package domain

import (
	"errors"
	"strings"
	"unicode"
)

const (
	PasswordMinLength = 8
	PasswordMaxLength = 128
)

var (
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong    = errors.New("password must be at most 128 characters")
	ErrPasswordNoUpper    = errors.New("password must contain an uppercase letter")
	ErrPasswordNoLower    = errors.New("password must contain a lowercase letter")
	ErrPasswordNoDigit    = errors.New("password must contain a digit")
	ErrPasswordCommon     = errors.New("password is too common; please choose a more unique one")
)

var commonPasswords = map[string]struct{}{
	"password":    {},
	"password1":   {},
	"password12":  {},
	"password123": {},
	"12345678":    {},
	"123456789":   {},
	"qwerty":      {},
	"qwerty123":   {},
	"admin":       {},
	"admin123":    {},
	"admin1234":   {},
	"admin12345":  {},
	"welcome":     {},
	"welcome1":    {},
	"welcome123":  {},
	"welcome1!":   {},
	"welc0me!":    {},
	"letmein":     {},
	"iloveyou":    {},
	"monkey":      {},
	"dragon":      {},
	"baseball":    {},
	"abc123":      {},
	"passw0rd":    {},
	"shadow":      {},
	"master":      {},
	"michael":     {},
	"superman":    {},
	"batman":      {},
	"trustno1":    {},
}

type Password struct {
	value string
}

func NewPassword(plain string) (Password, error) {
	trimmed := strings.TrimSpace(plain)
	if len(trimmed) < PasswordMinLength {
		return Password{}, ErrPasswordTooShort
	}
	if len(trimmed) > PasswordMaxLength {
		return Password{}, ErrPasswordTooLong
	}
	if isCommon(trimmed) {
		return Password{}, ErrPasswordCommon
	}
	if !hasVariety(trimmed) {
		return Password{}, ErrPasswordNoUpper
	}
	return Password{value: trimmed}, nil
}

func NewPasswordFromHash(hashed string) Password {
	return Password{value: hashed}
}

func (p Password) String() string {
	return p.value
}

func (p Password) Equals(other Password) bool {
	return p.value == other.value
}

func (p Password) IsEmpty() bool {
	return p.value == ""
}

func hasVariety(s string) bool {
	var hasUpper, hasLower, hasDigit bool
	for _, r := range s {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasUpper {
		return false
	}
	if !hasLower && !hasDigit {
		return false
	}
	return true
}

func isCommon(s string) bool {
	lower := strings.ToLower(s)
	_, found := commonPasswords[lower]
	return found
}
