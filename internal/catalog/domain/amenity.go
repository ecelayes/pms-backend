package domain

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

var ErrInvalidName = errors.New("invalid name")

type Amenity struct {
	id          string
	name        string
	description string
	icon        string
	createdAt   time.Time
}

func NewAmenity(name, description, icon string) (*Amenity, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	return &Amenity{
		id:          uuid.New().String(),
		name:        name,
		description: description,
		icon:        icon,
		createdAt:   time.Now(),
	}, nil
}
func (a *Amenity) ID() string          { return a.id }
func (a *Amenity) Name() string        { return a.name }
func (a *Amenity) Description() string { return a.description }
func (a *Amenity) Icon() string        { return a.icon }
func (a *Amenity) Update(name, description, icon string) {
	if name != "" {
		a.name = name
	}
	if description != "" {
		a.description = description
	}
	if icon != "" {
		a.icon = icon
	}
}
func ReconstituteAmenity(id, name, description, icon string) *Amenity {
	return &Amenity{
		id:          id,
		name:        name,
		description: description,
		icon:        icon,
		createdAt:   time.Now(),
	}
}
