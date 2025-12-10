package bootstrap

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/ecelayes/pms-backend/internal/handler"
	"github.com/ecelayes/pms-backend/internal/repository"
	"github.com/ecelayes/pms-backend/internal/security"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

func NewApp(pool *pgxpool.Pool) *echo.Echo {
	// 1. Repositories
	roomRepo := repository.NewRoomRepository(pool)
	resRepo := repository.NewReservationRepository(pool)
	hotelRepo := repository.NewHotelRepository(pool)
	priceRepo := repository.NewPriceRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	orgRepo := repository.NewOrganizationRepository(pool)
	guestRepo := repository.NewGuestRepository(pool)
	amenityRepo := repository.NewAmenityRepository(pool)
	serviceRepo := repository.NewHotelServiceRepository(pool)

	// 2. UseCases
	availUC := usecase.NewAvailabilityUseCase(roomRepo, resRepo, priceRepo)
	resUC := usecase.NewReservationUseCase(pool, roomRepo, resRepo, guestRepo)
	pricingUC := usecase.NewPricingUseCase(priceRepo, roomRepo)
	authUC := usecase.NewAuthUseCase(pool, userRepo)
	orgUC := usecase.NewOrganizationUseCase(orgRepo)
	userUC := usecase.NewUserUseCase(pool, userRepo, orgRepo)
	hotelUC := usecase.NewHotelUseCase(hotelRepo)
	roomUC := usecase.NewRoomUseCase(roomRepo)
	catalogUC := usecase.NewCatalogUseCase(amenityRepo, serviceRepo)

	// 3. Handlers
	availHandler := handler.NewAvailabilityHandler(availUC)
	resHandler := handler.NewReservationHandler(resUC)
	pricingHandler := handler.NewPricingHandler(pricingUC)
	authHandler := handler.NewAuthHandler(authUC)
	hotelHandler := handler.NewHotelHandler(hotelUC)
	roomHandler := handler.NewRoomHandler(roomUC)
	orgHandler := handler.NewOrganizationHandler(orgUC)
	userHandler := handler.NewUserHandler(userUC)
	catalogHandler := handler.NewCatalogHandler(catalogUC)

	// 4. Server Setup
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// 5. Routs
	v1 := e.Group("/api/v1")

	// Public
	v1.POST("/auth/login", authHandler.Login)
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
	protected.DELETE("/reservations/:id", resHandler.Delete, security.RequireSuperAdmin)

	// Users
	protected.POST("/users", userHandler.Create)
	protected.GET("/users", userHandler.GetAll)
	protected.GET("/users/:id", userHandler.GetByID)
	protected.PUT("/users/:id", userHandler.Update)
	protected.DELETE("/users/:id", userHandler.Delete)

	// Hotels CRUD
	protected.POST("/hotels", hotelHandler.Create)
	protected.GET("/hotels", hotelHandler.GetAll)
	protected.GET("/hotels/:id", hotelHandler.GetByID)
	protected.PUT("/hotels/:id", hotelHandler.Update)
	protected.DELETE("/hotels/:id", hotelHandler.Delete)

	// Room Types CRUD
	protected.POST("/room-types", roomHandler.Create)
	protected.GET("/room-types", roomHandler.GetAll)
	protected.GET("/room-types/:id", roomHandler.GetByID)
	protected.PUT("/room-types/:id", roomHandler.Update)
	protected.DELETE("/room-types/:id", roomHandler.Delete)

	// Pricing CRUD
	protected.POST("/pricing/rules", pricingHandler.CreateRule)
	protected.GET("/pricing/rules", pricingHandler.GetRules)
	protected.PUT("/pricing/rules/:id", pricingHandler.UpdateRule)
	protected.DELETE("/pricing/rules/:id", pricingHandler.DeleteRule)

	return e
}
