package domain

import (
	"github.com/google/uuid"
	"time"
)

type GuestService struct {
	id          string
	name        string
	description string
	icon        string
	createdAt   time.Time
}

func NewGuestService(name, description, icon string) (*GuestService, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	return &GuestService{
		id:          uuid.New().String(),
		name:        name,
		description: description,
		icon:        icon,
		createdAt:   time.Now(),
	}, nil
}
func (s *GuestService) ID() string          { return s.id }
func (s *GuestService) Name() string        { return s.name }
func (s *GuestService) Description() string { return s.description }
func (s *GuestService) Icon() string        { return s.icon }
func (s *GuestService) Update(name, description, icon string) {
	if name != "" {
		s.name = name
	}
	if description != "" {
		s.description = description
	}
	if icon != "" {
		s.icon = icon
	}
}
func ReconstituteGuestService(id, name, description, icon string) *GuestService {
	return &GuestService{
		id:          id,
		name:        name,
		description: description,
		icon:        icon,
		createdAt:   time.Now(),
	}
}
