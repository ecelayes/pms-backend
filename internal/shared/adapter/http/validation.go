package http

import (
	"errors"
	"net/mail"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidEmail   = errors.New("invalid email format")
	ErrRequiredField  = errors.New("field is required")
	ErrInvalidUUID    = errors.New("invalid UUID format")
)

func ValidateEmail(s string) error {
	parsed, err := mail.ParseAddress(s)
	if err != nil {
		return ErrInvalidEmail
	}
	// Reject addresses with display names: "John Doe <john@x.com>" — too lenient
	if parsed.Name != "" {
		return ErrInvalidEmail
	}
	return nil
}

func ValidateRequired(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(name + " " + ErrRequiredField.Error())
	}
	return nil
}

func ValidateUUID(s string) error {
	if s == "" {
		return ErrInvalidUUID
	}
	if _, err := uuid.Parse(s); err != nil {
		return ErrInvalidUUID
	}
	return nil
}
