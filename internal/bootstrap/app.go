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
	guestRepo := repository.NewGuestRepository(pool)

	// 2. UseCases
	availUC := usecase.NewAvailabilityUseCase(roomRepo)
	resUC := usecase.NewReservationUseCase(pool, roomRepo, resRepo, guestRepo)
	pricingUC := usecase.NewPricingUseCase(priceRepo, roomRepo)
	authUC := usecase.NewAuthUseCase(userRepo)
	hotelUC := usecase.NewHotelUseCase(hotelRepo)
	roomUC := usecase.NewRoomUseCase(roomRepo)

	// 3. Handlers
	availHandler := handler.NewAvailabilityHandler(availUC)
	resHandler := handler.NewReservationHandler(resUC)
	pricingHandler := handler.NewPricingHandler(pricingUC)
	authHandler := handler.NewAuthHandler(authUC)
	hotelHandler := handler.NewHotelHandler(hotelUC)
	roomHandler := handler.NewRoomHandler(roomUC)

	// 4. Server Setup
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// 5. Routs
	v1 := e.Group("/api/v1")

	// Public
	v1.POST("/auth/register", authHandler.Register)
	v1.POST("/auth/login", authHandler.Login)
	v1.GET("/availability", availHandler.Get)
	v1.POST("/reservations", resHandler.Create)
	v1.GET("/reservations/:code", resHandler.GetByCode)
	v1.POST("/reservations/:id/cancel", resHandler.Cancel)

	// Protected
	protected := v1.Group("")
	protected.Use(security.Auth(authUC))
	
	// Hotels CRUD
	protected.POST("/hotels", hotelHandler.Create)
	protected.GET("/hotels", hotelHandler.ListMine)
	protected.PUT("/hotels/:id", hotelHandler.Update)
	protected.DELETE("/hotels/:id", hotelHandler.Delete)

	// Room Types CRUD
	protected.POST("/room-types", roomHandler.Create)
	protected.PUT("/room-types/:id", roomHandler.Update)
	protected.DELETE("/room-types/:id", roomHandler.Delete)

	// Pricing CRUD
	protected.POST("/pricing/rules", pricingHandler.CreateRule)
	protected.PUT("/pricing/rules/:id", pricingHandler.UpdateRule)
	protected.DELETE("/pricing/rules/:id", pricingHandler.DeleteRule)

	// Reservation Admin
	protected.DELETE("/reservations/:id", resHandler.Delete)

	return e
}
