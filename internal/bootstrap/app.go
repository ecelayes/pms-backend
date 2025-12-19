package bootstrap

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/ecelayes/pms-backend/pkg/logger"
	"github.com/ecelayes/pms-backend/internal/handler"
	"github.com/ecelayes/pms-backend/internal/repository"
	"github.com/ecelayes/pms-backend/internal/security"
	"github.com/ecelayes/pms-backend/internal/usecase"
	"github.com/ecelayes/pms-backend/internal/service"
)

func NewApp(pool *pgxpool.Pool) *echo.Echo {
	// 0. Logger
	log, err := logger.New()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	// 1. Repositories
	unitTypeRepo := repository.NewUnitTypeRepository(pool)
	unitRepo := repository.NewUnitRepository(pool)
	resRepo := repository.NewReservationRepository(pool)
	propertyRepo := repository.NewPropertyRepository(pool)
	priceRepo := repository.NewPriceRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	orgRepo := repository.NewOrganizationRepository(pool)
	guestRepo := repository.NewGuestRepository(pool)
	amenityRepo := repository.NewAmenityRepository(pool)
	serviceRepo := repository.NewHotelServiceRepository(pool)
	ratePlanRepo := repository.NewRatePlanRepository(pool)

	// 1.5 Domain Services
	pricingService := service.NewPricingService(priceRepo)
	inventoryService := service.NewInventoryService()
	emailService := service.NewEmailService()

	// 2. UseCases
	availUC := usecase.NewAvailabilityUseCase(unitTypeRepo, resRepo, ratePlanRepo, pricingService)
	resUC := usecase.NewReservationUseCase(pool, unitTypeRepo, resRepo, guestRepo, ratePlanRepo, pricingService)
	pricingUC := usecase.NewPricingUseCase(pool, priceRepo, unitTypeRepo, inventoryService)
	authUC := usecase.NewAuthUseCase(pool, userRepo, orgRepo, emailService, log)
	orgUC := usecase.NewOrganizationUseCase(orgRepo)
	userUC := usecase.NewUserUseCase(pool, userRepo, orgRepo)
	propertyUC := usecase.NewPropertyUseCase(propertyRepo)
	unitTypeUC := usecase.NewUnitTypeUseCase(unitTypeRepo)
	unitUC := usecase.NewUnitUseCase(unitRepo)
	catalogUC := usecase.NewCatalogUseCase(amenityRepo, serviceRepo)
	ratePlanUC := usecase.NewRatePlanUseCase(ratePlanRepo, resRepo)

	// 3. Handlers
	availHandler := handler.NewAvailabilityHandler(availUC)
	resHandler := handler.NewReservationHandler(resUC)
	pricingHandler := handler.NewPricingHandler(pricingUC)
	authHandler := handler.NewAuthHandler(authUC)
	propertyHandler := handler.NewPropertyHandler(propertyUC)
	unitTypeHandler := handler.NewUnitTypeHandler(unitTypeUC)
	unitHandler := handler.NewUnitHandler(unitUC)
	orgHandler := handler.NewOrganizationHandler(orgUC)
	userHandler := handler.NewUserHandler(userUC)
	catalogHandler := handler.NewCatalogHandler(catalogUC)
	ratePlanHandler := handler.NewRatePlanHandler(ratePlanUC)

	// 4. Server Setup
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("logger", log)
			return next(c)
		}
	})

	// 5. Routs
	v1 := e.Group("/api/v1")

	// Public
	v1.POST("/auth/login", authHandler.Login)
	v1.POST("/auth/forgot-password", authHandler.ForgotPassword)
	v1.POST("/auth/reset-password", authHandler.ResetPassword)
	v1.GET("/availability", availHandler.Get)
	v1.POST("/reservations", resHandler.Create)
	v1.GET("/reservations/:code", resHandler.GetByCode)
	v1.POST("/reservations/:id/cancel", resHandler.Cancel)

	// Protected
	protected := v1.Group("")
	protected.Use(security.Auth(authUC))

	// Organizations
	protected.POST("/organizations", orgHandler.Create, security.RequireSuperAdmin)
	protected.GET("/organizations", orgHandler.GetAll, security.RequireSuperAdmin)
	protected.GET("/organizations/:id", orgHandler.GetByID, security.RequireSuperAdmin)
	protected.PUT("/organizations/:id", orgHandler.Update, security.RequireSuperAdmin)
	protected.DELETE("/organizations/:id", orgHandler.Delete, security.RequireSuperAdmin)

	// Amenities CRUD
	protected.POST("/amenities", catalogHandler.CreateAmenity, security.RequireSuperAdmin)
	protected.PUT("/amenities/:id", catalogHandler.UpdateAmenity, security.RequireSuperAdmin)
	protected.DELETE("/amenities/:id", catalogHandler.DeleteAmenity, security.RequireSuperAdmin)
	protected.GET("/amenities", catalogHandler.GetAllAmenities) 
	protected.GET("/amenities/:id", catalogHandler.GetAmenityByID) 

	// Services CRUD
	protected.POST("/services", catalogHandler.CreateService, security.RequireSuperAdmin)
	protected.PUT("/services/:id", catalogHandler.UpdateService, security.RequireSuperAdmin)
	protected.DELETE("/services/:id", catalogHandler.DeleteService, security.RequireSuperAdmin)
	protected.GET("/services", catalogHandler.GetAllServices)
	protected.GET("/services/:id", catalogHandler.GetServiceByID)

	// Reservation Admin
	protected.GET("/reservations/:id/cancel-preview", resHandler.PreviewCancel)
	protected.DELETE("/reservations/:id", resHandler.Delete, security.RequireSuperAdmin)

	// Users
	protected.POST("/users", userHandler.Create)
	protected.GET("/users", userHandler.GetAll)
	protected.GET("/users/:id", userHandler.GetByID)
	protected.PUT("/users/:id", userHandler.Update)
	protected.DELETE("/users/:id", userHandler.Delete)

	// Properties CRUD
	protected.POST("/properties", propertyHandler.Create)
	protected.GET("/properties", propertyHandler.GetAll)
	protected.GET("/properties/:id", propertyHandler.GetByID)
	protected.PUT("/properties/:id", propertyHandler.Update)
	protected.DELETE("/properties/:id", propertyHandler.Delete)

	// Unit Types CRUD
	protected.POST("/unit-types", unitTypeHandler.Create)
	protected.GET("/unit-types", unitTypeHandler.GetAll)
	protected.GET("/unit-types/:id", unitTypeHandler.GetByID)
	protected.PUT("/unit-types/:id", unitTypeHandler.Update)
	protected.DELETE("/unit-types/:id", unitTypeHandler.Delete)

	// Units CRUD
	protected.POST("/units", unitHandler.Create)
	protected.GET("/units", unitHandler.GetAll)
	protected.GET("/units/:id", unitHandler.GetByID)
	protected.PUT("/units/:id", unitHandler.Update)
	protected.DELETE("/units/:id", unitHandler.Delete)

	// Pricing CRUD
	protected.POST("/pricing/bulk", pricingHandler.BulkUpdate)
	protected.GET("/pricing/rules", pricingHandler.GetRules)
	protected.DELETE("/pricing/rules/:id", pricingHandler.DeleteRule)

	// Rate Plans CRUD
	protected.POST("/rate-plans", ratePlanHandler.Create)
	protected.GET("/rate-plans", ratePlanHandler.List)
	protected.GET("/rate-plans/:id", ratePlanHandler.GetByID)
	protected.PUT("/rate-plans/:id", ratePlanHandler.Update)
	protected.DELETE("/rate-plans/:id", ratePlanHandler.Delete)

	return e
}
