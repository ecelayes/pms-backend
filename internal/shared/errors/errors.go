package errors

import "errors"

// Standard domain errors used across all bounded contexts
var (
	// ErrNotFound is returned when an entity cannot be found in the repository
	ErrNotFound = errors.New("record not found")
	
	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")
	
	// ErrUnauthorized is returned when an operation is not permitted
	ErrUnauthorized = errors.New("unauthorized")
	
	// ErrConflict is returned when there's a conflict with existing data
	ErrConflict = errors.New("conflict")
	
	// ErrInternal is returned for unexpected internal errors
	ErrInternal = errors.New("internal error")
)
