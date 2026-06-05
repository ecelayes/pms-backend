package domain

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

var ErrInvalidOrganization = errors.New("invalid organization")

type Organization struct {
	id        string
	name      string
	code      string
	createdAt time.Time
}

func NewOrganization(name, code string) (*Organization, error) {
	if name == "" || code == "" {
		return nil, ErrInvalidOrganization
	}
	return &Organization{
		id:        uuid.New().String(),
		name:      name,
		code:      code,
		createdAt: time.Now(),
	}, nil
}
func ReconstituteOrganization(id, name, code string, createdAt time.Time) *Organization {
	return &Organization{
		id:        id,
		name:      name,
		code:      code,
		createdAt: createdAt,
	}
}
func (o *Organization) ID() string   { return o.id }
func (o *Organization) Name() string { return o.name }
func (o *Organization) Code() string { return o.code }
func (o *Organization) Update(name, code string) {
	if name != "" {
		o.name = name
	}
	if code != "" {
		o.code = code
	}
}

type OrganizationMember struct {
}
