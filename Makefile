# --- VARS ---
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

APP_NAME=pms-backend
TARGET_DB ?= $(DB_NAME)
GO_TEST_DSN := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_TEST_NAME)?sslmode=disable

.PHONY: help run build test clean db-up db-seed db-reset test-prepare test-all test-unit docker-up docker-down docker-db-reset \
        test-lifecycle test-hotel test-auth test-room test-pricing test-reservation

help:
	@echo 'Usage: make [target]'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# --- APP ---
run:
	@echo "Starting $(APP_NAME)"
	go run cmd/api/main.go

build:
	go build -o bin/$(APP_NAME) cmd/api/main.go

clean:
	rm -rf bin/
	go clean

# --- DATABASE ---
db-up:
	@echo "Applying schema to $(TARGET_DB)"
	@cat migrations/*.up.sql | sudo docker compose exec -T db psql -U $(DB_USER) -d $(TARGET_DB)

db-seed:
	@echo "Seeding data to $(TARGET_DB)"
	@cat scripts/seed_data.sql | sudo docker compose exec -T db psql -U $(DB_USER) -d $(TARGET_DB)

db-reset:
	@echo "Resetting Database $(DB_NAME)"
	sudo docker compose exec -T db psql -U $(DB_USER) -d $(DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@make db-up
	@make db-seed
	@echo "Database Ready!"

# --- TESTING ---
test-prepare:
	@echo "Preparing Test Database: $(DB_TEST_NAME)"
	@sudo docker compose exec -T db psql -U $(DB_USER) -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$(DB_TEST_NAME)' AND pid <> pg_backend_pid();" || true
	@sudo docker compose exec -T db psql -U $(DB_USER) -d postgres -c "DROP DATABASE IF EXISTS $(DB_TEST_NAME);"
	@sudo docker compose exec -T db psql -U $(DB_USER) -d postgres -c "CREATE DATABASE $(DB_TEST_NAME);"
	@make db-up TARGET_DB=$(DB_TEST_NAME)

test-all: test-prepare
	@echo "Running ALL Tests"
	export TEST_DATABASE_URL="$(GO_TEST_DSN)"; go test -v ./tests/... -skip TestLifecycleSuite
	export TEST_DATABASE_URL="$(GO_TEST_DSN)"; go test -v ./tests/... -run TestLifecycleSuite

test-unit: test-prepare
	@echo "Running Unit Tests"
	export TEST_DATABASE_URL="$(GO_TEST_DSN)"; go test -v ./tests/... -skip TestLifecycleSuite

test-lifecycle: test-prepare
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

test-reservation: test-prepare
	@echo "Running Reservation Tests"
	export TEST_DATABASE_URL="$(GO_TEST_DSN)"; go test -v ./tests/... -run TestReservationSuite

# --- DOCKER ---
docker-up:
	sudo docker compose up -d --build

docker-down:
	sudo docker compose down -v

docker-db-reset:
	@echo "Resetting Docker DB"
	@sleep 2
	sudo docker compose exec -T db psql -U $(DB_USER) -d $(DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	cat migrations/*.up.sql | sudo docker compose exec -T db psql -U $(DB_USER) -d $(DB_NAME)
	cat scripts/seed_data.sql | sudo docker compose exec -T db psql -U $(DB_USER) -d $(DB_NAME)
