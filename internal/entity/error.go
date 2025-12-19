package entity

import "errors"

var (
	// Input Validation
	ErrInvalidDateFormat = errors.New("invalid date format (use YYYY-MM-DD)")
	ErrInvalidDateRange  = errors.New("check-out date must be after check-in date")
	ErrInvalidID         = errors.New("invalid UUID format")
	ErrEmptyBody         = errors.New("request body cannot be empty")
	ErrInvalidInput      = errors.New("invalid input")
	ErrConflict          = errors.New("conflict: resource already exists")

	// Business Rules (Booking)
	ErrNoAvailability       = errors.New("no availability for selected dates")
	ErrReservationNotFound  = errors.New("reservation not found")
	ErrReservationCancelled = errors.New("reservation is already cancelled")
	
	// Business Rules (Pricing)
	ErrPriceNegative 		= errors.New("price must be positive")
	ErrPriorityNegative = errors.New("priority cannot be negative")

	// Integrity
	ErrUnitTypeNotFound = errors.New("unit type ID does not exist")

	// Auth
	ErrEmailAlreadyExists = errors.New("email is already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user account is inactive")

	// Permissions
	ErrInsufficientPermissions = errors.New("insufficient permissions")

	// General
	ErrRecordNotFound = errors.New("record not found")
	ErrInternal       = errors.New("internal server error")
)
