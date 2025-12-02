package entity

import "errors"

var (
	// 400 Bad Request
	ErrInvalidDateRange = errors.New("check-out date must be after check-in date")
	ErrInvalidDateFormat = errors.New("invalid date format (use YYYY-MM-DD)")
	
	// 409 Conflict
	ErrNoAvailability = errors.New("no availability for selected dates")
	ErrPriceChanged   = errors.New("price has changed during booking")
	
	// 404 Not Found
	ErrRoomTypeNotFound = errors.New("room type ID does not exist")
)
