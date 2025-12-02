package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ecelayes/pms-backend/internal/handler"
	"github.com/ecelayes/pms-backend/internal/repository"
	"github.com/ecelayes/pms-backend/internal/usecase"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	roomRepo := repository.NewRoomRepository(pool)

	availUseCase := usecase.NewAvailabilityUseCase(roomRepo)
	resUseCase := usecase.NewReservationUseCase(pool, roomRepo)

	availHandler := handler.NewAvailabilityHandler(availUseCase)
	resHandler := handler.NewReservationHandler(resUseCase)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	v1 := e.Group("/api/v1")
	v1.GET("/availability", availHandler.Get)
	v1.POST("/reservations", resHandler.Create)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
