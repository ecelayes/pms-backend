package domain

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("record not found")

type PropertyRepository interface {
	SaveProperty(ctx context.Context, p *Property) error
	FindPropertyByID(ctx context.Context, id string) (*Property, error)
	FindPropertiesByOrganization(ctx context.Context, orgID string, limit, offset int) ([]*Property, int64, error)
	DeleteProperty(ctx context.Context, id string) error
}
type UnitTypeRepository interface {
	SaveUnitType(ctx context.Context, ut *UnitType) error
	FindUnitTypeByID(ctx context.Context, id string) (*UnitType, error)
	FindUnitTypesByPropertyID(ctx context.Context, propertyID string, limit, offset int) ([]*UnitType, int64, error)
	DeleteUnitType(ctx context.Context, id string) error
}
type UnitRepository interface {
	SaveUnit(ctx context.Context, u *Unit) error
	FindUnitByID(ctx context.Context, id string) (*Unit, error)
	FindUnitsByPropertyID(ctx context.Context, propertyID string, limit, offset int) ([]*Unit, int64, error)
	DeleteUnit(ctx context.Context, id string) error
}
type AmenityRepository interface {
	Save(ctx context.Context, a *Amenity) error
	FindAll(ctx context.Context, limit, offset int) ([]*Amenity, int64, error)
	FindByID(ctx context.Context, id string) (*Amenity, error)
	Delete(ctx context.Context, id string) error
}
type GuestServiceRepository interface {
	SaveGuestService(ctx context.Context, gs *GuestService) error
	FindAllGuestServices(ctx context.Context, limit, offset int) ([]*GuestService, int64, error)
	FindGuestServiceByID(ctx context.Context, id string) (*GuestService, error)
	DeleteGuestService(ctx context.Context, id string) error
}
