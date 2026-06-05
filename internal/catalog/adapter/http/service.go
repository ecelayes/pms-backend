package http

import (
	"context"

	"github.com/ecelayes/pms-backend/internal/catalog/application"
	"github.com/ecelayes/pms-backend/internal/catalog/domain"
)

// PropertyService defines property operations exposed to the HTTP layer.
type PropertyService interface {
	CreateProperty(ctx context.Context, orgID, name, code, pTypeStr string) (string, error)
	GetProperty(ctx context.Context, id string) (*domain.Property, error)
	ListProperties(ctx context.Context, orgID string, page, limit int) ([]*domain.Property, int64, error)
	UpdateProperty(ctx context.Context, id, name, code, pTypeStr string) error
	DeleteProperty(ctx context.Context, id string) error
}

// UnitTypeService defines unit type operations.
type UnitTypeService interface {
	CreateUnitType(ctx context.Context, propertyID, name, code string, qty int, priceCents int64, currency string, maxOcc, maxAd, maxCh int, amenities []string) (string, error)
	ListUnitTypes(ctx context.Context, propertyID string, page, limit int) ([]*domain.UnitType, int64, error)
	GetUnitTypeEntity(ctx context.Context, id string) (*domain.UnitType, error)
	UpdateUnitType(ctx context.Context, id, name, code string, qty int, priceCents int64, currency string, maxOcc, maxAd, maxCh int, amenities []string) error
	DeleteUnitType(ctx context.Context, id string) error
}

// UnitService defines unit operations.
type UnitService interface {
	CreateUnit(ctx context.Context, propertyID, unitTypeID, name string) (string, error)
	GetUnit(ctx context.Context, id string) (*domain.Unit, error)
	ListUnits(ctx context.Context, propertyID string, page, limit int) ([]*domain.Unit, int64, error)
	UpdateUnit(ctx context.Context, id, name, status string) error
	DeleteUnit(ctx context.Context, id string) error
}

// CatalogService defines catalog item operations (amenities + guest services).
type CatalogService interface {
	CreateAmenity(ctx context.Context, name, description, icon string) (string, error)
	GetAmenity(ctx context.Context, id string) (*domain.Amenity, error)
	ListAmenities(ctx context.Context, page, limit int) ([]*domain.Amenity, int64, error)
	UpdateAmenity(ctx context.Context, id, name, description, icon string) error
	DeleteAmenity(ctx context.Context, id string) error
	CreateGuestService(ctx context.Context, name, description, icon string) (string, error)
	GetGuestService(ctx context.Context, id string) (*domain.GuestService, error)
	ListGuestServices(ctx context.Context, page, limit int) ([]*domain.GuestService, int64, error)
	UpdateGuestService(ctx context.Context, id, name, description, icon string) error
	DeleteGuestService(ctx context.Context, id string) error
}

// Verify that the concrete *application.CatalogService implements all four interfaces.
var (
	_ PropertyService = (*application.CatalogService)(nil)
	_ UnitTypeService = (*application.CatalogService)(nil)
	_ UnitService     = (*application.CatalogService)(nil)
	_ CatalogService  = (*application.CatalogService)(nil)
)
