package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ecelayes/pms-backend/internal/bootstrap"
	"github.com/ecelayes/pms-backend/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	defaultReadHeaderTimeout = 10 * time.Second
	defaultReadTimeout       = 30 * time.Second
	defaultWriteTimeout      = 30 * time.Second
	defaultIdleTimeout       = 120 * time.Second
	defaultShutdownTimeout   = 15 * time.Second
)

func main() {
	_ = godotenv.Load()

	appLogger, err := logger.New()
	if err != nil {
		log.Fatalf("Unable to initialize logger: %v", err)
	}
	defer func() { _ = appLogger.Sync() }()

	dbURL := os.Getenv("DATABASE_URL")
	dbConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		appLogger.Fatal("unable to parse database URL", zap.Error(err))
	}
	dbConfig.MaxConns = 25
	dbConfig.MinConns = 5
	dbConfig.MaxConnLifetime = 1 * time.Hour
	dbConfig.HealthCheckPeriod = 1 * time.Minute
	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		appLogger.Fatal("unable to connect to database", zap.Error(err))
	}
	defer pool.Close()

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		appLogger.Warn("unable to connect to Redis", zap.String("addr", redisAddr), zap.Error(err))
	}
	defer func() {
		if err := rdb.Close(); err != nil {
			appLogger.Warn("failed to close redis", zap.Error(err))
		}
	}()

	e := bootstrap.NewApp(pool, rdb, appLogger)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           e,
		ReadHeaderTimeout: getEnvDuration("SERVER_READ_HEADER_TIMEOUT", defaultReadHeaderTimeout),
		ReadTimeout:       getEnvDuration("SERVER_READ_TIMEOUT", defaultReadTimeout),
		WriteTimeout:      getEnvDuration("SERVER_WRITE_TIMEOUT", defaultWriteTimeout),
		IdleTimeout:       getEnvDuration("SERVER_IDLE_TIMEOUT", defaultIdleTimeout),
	}

	go func() {
		appLogger.Info("server starting", zap.String("port", port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	appLogger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", defaultShutdownTimeout))
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("graceful shutdown failed", zap.Error(err))
	}
	appLogger.Info("server stopped")
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	if n, err := strconv.Atoi(v); err == nil {
		return time.Duration(n) * time.Second
	}
	return def
}
