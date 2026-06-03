package domain

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

var (
	ErrRequestInvalid = errors.New("request invalid")
)

type PropertyType string

const (
	PropertyTypeHotel     PropertyType = "hotel"
	PropertyTypeApartment PropertyType = "apartment"
	PropertyTypeCabin     PropertyType = "cabin"
)

type Property struct {
	id             string
	organizationID string
	name           string
	code           string
	propertyType   PropertyType
	createdAt      time.Time
}

func NewProperty(organizationID, name, code string, pType PropertyType) (*Property, error) {
	if name == "" || code == "" {
		return nil, ErrRequestInvalid
	}
	return &Property{
		id:             uuid.New().String(),
		organizationID: organizationID,
		name:           name,
		code:           code,
		propertyType:   pType,
		createdAt:      time.Now(),
	}, nil
}
func ReconstituteProperty(id, orgID, name, code, pTypeStr string, createdAt time.Time) *Property {
	return &Property{
		id:             id,
		organizationID: orgID,
		name:           name,
		code:           code,
		propertyType:   PropertyType(pTypeStr),
		createdAt:      createdAt,
	}
}
func (p *Property) ID() string             { return p.id }
func (p *Property) OrganizationID() string { return p.organizationID }
func (p *Property) Name() string           { return p.name }
func (p *Property) Code() string           { return p.code }
func (p *Property) Type() PropertyType     { return p.propertyType }
func (p *Property) Update(name, code string, pType PropertyType) {
	if name != "" {
		p.name = name
	}
	if code != "" {
		p.code = code
	}
	if pType != "" {
		p.propertyType = pType
	}
}
