package catalog

import (
	"github.com/ecelayes/pms-backend/internal/catalog/adapter"
	"github.com/ecelayes/pms-backend/internal/catalog/adapter/http"
	"github.com/ecelayes/pms-backend/internal/catalog/application"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type Module struct {
	Service *application.CatalogService
}

func NewModule(db *pgxpool.Pool, group *echo.Group, protected *echo.Group) *Module {
	repo := adapter.NewPostgresCatalogRepository(db)
	amenityRepo := adapter.NewPostgresAmenityRepository(db)
	guestServiceRepo := adapter.NewPostgresGuestServiceRepository(db)
	service := application.NewCatalogService(repo, repo, repo, amenityRepo, guestServiceRepo)
	catalogH := http.NewCatalogHandler(service)
	propH := http.NewPropertyHandler(service)
	utH := http.NewUnitTypeHandler(service)
	unitH := http.NewUnitHandler(service)
	protected.POST("/amenities", catalogH.CreateAmenity)
	protected.PUT("/amenities/:id", catalogH.UpdateAmenity)
	protected.DELETE("/amenities/:id", catalogH.DeleteAmenity)
	protected.GET("/amenities", catalogH.GetAllAmenities)
	protected.GET("/amenities/:id", catalogH.GetAmenityByID)
	protected.POST("/services", catalogH.CreateService)
	protected.PUT("/services/:id", catalogH.UpdateService)
	protected.DELETE("/services/:id", catalogH.DeleteService)
	protected.GET("/services", catalogH.GetAllServices)
	protected.GET("/services/:id", catalogH.GetServiceByID)
	protected.POST("/properties", propH.Create)
	protected.GET("/properties", propH.GetAll)
	protected.GET("/properties/:id", propH.GetByID)
	protected.PUT("/properties/:id", propH.Update)
	protected.DELETE("/properties/:id", propH.Delete)
	protected.POST("/unit-types", utH.Create)
	protected.GET("/unit-types", utH.GetAll)
	protected.GET("/unit-types/:id", utH.GetByID)
	protected.PUT("/unit-types/:id", utH.Update)
	protected.DELETE("/unit-types/:id", utH.Delete)
	protected.POST("/units", unitH.Create)
	protected.GET("/units", unitH.GetAll)
	protected.GET("/units/:id", unitH.GetByID)
	protected.PUT("/units/:id", unitH.Update)
	protected.DELETE("/units/:id", unitH.Delete)
	return &Module{Service: service}
}
