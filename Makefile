# --- VARS ---
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

APP_NAME=pms-backend
TARGET_DB ?= $(DB_NAME)
GO_TEST_DSN := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_TEST_NAME)?sslmode=disable

.PHONY: help run build test clean db-up db-seed db-reset test-prepare test-all test-unit docker-up docker-down docker-db-reset \
        prod-up prod-down prod-logs \
        test-lifecycle test-hotel test-auth test-room test-pricing test-reservation

help:
	@echo 'Usage: make [target]'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# --- APP ---
run: ## Run API locally (go run)
	@echo "Starting $(APP_NAME)"
	go run cmd/api/main.go

build: ## Build binary
	go build -o bin/$(APP_NAME) cmd/api/main.go

clean: ## Remove binary
	rm -rf bin/
	go clean

# --- DATABASE (Local/Hybrid) ---
db-up: ## Apply migrations to DB container
	@echo "Applying schema to $(TARGET_DB)"
	@cat migrations/*.up.sql | sudo docker compose -f docker-compose.yml exec -T db psql -U $(DB_USER) -d $(TARGET_DB)

db-seed: ## Seed data to DB container
	@echo "Seeding data to $(TARGET_DB)"
	@cat scripts/seed_data.sql | sudo docker compose -f docker-compose.yml exec -T db psql -U $(DB_USER) -d $(TARGET_DB)

db-reset: ## Full DB Reset (Drop + Up + Seed)
	@echo "Resetting Database $(DB_NAME)"
	sudo docker compose -f docker-compose.yml exec -T db psql -U $(DB_USER) -d $(DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@make db-up
	@make db-seed
	@echo "Database Ready!"

# --- TESTING ---
test-prepare:
	@echo "Preparing Test Database: $(DB_TEST_NAME)"
	@sudo docker compose -f docker-compose.yml exec -T db psql -U $(DB_USER) -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$(DB_TEST_NAME)' AND pid <> pg_backend_pid();" || true
	@sudo docker compose -f docker-compose.yml exec -T db psql -U $(DB_USER) -d postgres -c "DROP DATABASE IF EXISTS $(DB_TEST_NAME);"
	@sudo docker compose -f docker-compose.yml exec -T db psql -U $(DB_USER) -d postgres -c "CREATE DATABASE $(DB_TEST_NAME);"
	@make db-up TARGET_DB=$(DB_TEST_NAME)

test-all: test-prepare ## Run ALL integration tests
	@echo "Running ALL Tests"
	export TEST_DATABASE_URL="$(GO_TEST_DSN)"; go test -v ./tests/... -skip TestLifecycleSuite
	export TEST_DATABASE_URL="$(GO_TEST_DSN)"; go test -v ./tests/... -run TestLifecycleSuite

test-unit: test-prepare ## Run Unit tests only
	@echo "Running Unit Tests"
	export TEST_DATABASE_URL="$(GO_TEST_DSN)"; go test -v ./tests/... -skip TestLifecycleSuite

test-lifecycle: test-prepare ## Run Lifecycle test
	@echo "Running Lifecycle Test"
	export TEST_DATABASE_URL="$(GO_TEST_DSN)"; go test -v ./tests/... -run TestLifecycleSuite

test-auth: test-prepare
	@echo "Running Auth Tests"
	export TEST_DATABASE_URL="$(GO_TEST_DSN)"; go test -v ./tests/... -run TestAuthSuite

test-hotel: test-prepare
	@echo "Running Hotel Tests"
	export TEST_DATABASE_URL="$(GO_TEST_DSN)"; go test -v ./tests/... -run TestHotelSuite

test-room: test-prepare
	@echo "Running Room Tests"
	export TEST_DATABASE_URL="$(GO_TEST_DSN)"; go test -v ./tests/... -run TestRoomSuite

test-pricing: test-prepare
	@echo "Running Pricing Tests"
	export TEST_DATABASE_URL="$(GO_TEST_DSN)"; go test -v ./tests/... -run TestPricingSuite

test-rate-plan: test-prepare
	@echo "Running Rate Plan Tests"
	export TEST_DATABASE_URL="$(GO_TEST_DSN)"; go test -v ./tests/... -run TestRatePlanSuite

test-reservation: test-prepare
	@echo "Running Reservation Tests"
	export TEST_DATABASE_URL="$(GO_TEST_DSN)"; go test -v ./tests/... -run TestReservationSuite

# --- DOCKER DEV (Hot Reload) ---
docker-up: ## Start Dev Environment (Air + DB)
	sudo docker compose -f docker-compose.yml up -d --build

docker-down: ## Stop Dev Environment
	sudo docker compose -f docker-compose.yml down

docker-logs: ## Follow Dev Logs
	sudo docker compose -f docker-compose.yml logs -f

docker-db-reset: ## Reset Dev DB from scratch
	@echo "Resetting Docker Dev DB"
	@sleep 2
	sudo docker compose -f docker-compose.yml exec -T db psql -U $(DB_USER) -d $(DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	cat migrations/*.up.sql | sudo docker compose -f docker-compose.yml exec -T db psql -U $(DB_USER) -d $(DB_NAME)
	cat scripts/seed_data.sql | sudo docker compose -f docker-compose.yml exec -T db psql -U $(DB_USER) -d $(DB_NAME)

# --- DOCKER PROD ---
prod-up: ## Start Production Environment
	sudo docker compose -f docker-compose.prod.yml up -d --build

prod-down: ## Stop Production Environment
	sudo docker compose -f docker-compose.prod.yml down

prod-logs: ## Follow Production Logs
	sudo docker compose -f docker-compose.prod.yml logs -f
