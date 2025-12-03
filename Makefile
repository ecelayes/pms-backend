# --- VARS ---
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

APP_NAME=pms-backend

DB_DSN := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
TEST_DB_DSN := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_TEST_NAME)?sslmode=disable
DB_ADMIN_DSN := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/postgres?sslmode=disable

.PHONY: help run build test clean db-up db-seed db-reset test-prepare test-all test-unit test-lifecycle test-run docker-up docker-down docker-db-reset

help:
	@echo 'Usage: make [target]'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# --- APP ---
run:
	@echo "Starting $(APP_NAME)..."
	go run cmd/api/main.go

build:
	go build -o bin/$(APP_NAME) cmd/api/main.go

clean:
	rm -rf bin/
	go clean

# --- DATABASE ---
db-up:
	@echo "Applying schema to $(DB_NAME)..."
	@cat migrations/*.up.sql | psql "$(DB_DSN)"

db-seed:
	@echo "Seeding data..."
	@cat scripts/seed_data.sql | psql "$(DB_DSN)"

db-reset:
	@echo "Resetting Database $(DB_NAME)..."
	psql "$(DB_DSN)" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@make db-up
	@make db-seed
	@echo "Database Ready!"

# --- TESTING ---
test-prepare:
	@echo "Preparing Test Database: $(DB_TEST_NAME)..."
	@psql "$(DB_ADMIN_DSN)" -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$(DB_TEST_NAME)' AND pid <> pg_backend_pid();" || true
	@psql "$(DB_ADMIN_DSN)" -c "DROP DATABASE IF EXISTS $(DB_TEST_NAME);"
	@psql "$(DB_ADMIN_DSN)" -c "CREATE DATABASE $(DB_TEST_NAME);"
	@make db-up DB_DSN="$(TEST_DB_DSN)"

test-all: test-prepare
	@echo "Running ALL Tests..."
	export TEST_DATABASE_URL="$(TEST_DB_DSN)"; go test -v ./...

test-unit:
	@echo "Running Unit Tests..."
	go test -v ./internal/...

test-lifecycle: test-prepare
	@echo "Running Lifecycle E2E..."
	export TEST_DATABASE_URL="$(TEST_DB_DSN)"; go test -v ./tests/ -run TestLifecycleSuite

test-run: test-prepare
	@if [ -z "$(name)" ]; then echo "Error: define name. Ej: make test-run name=TestHotelSuite"; exit 1; fi
	@echo "Running Specific Test: $(name)..."
	export TEST_DATABASE_URL="$(TEST_DB_DSN)"; go test -v ./... -run $(name)

# --- DOCKER ---
docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

docker-db-reset:
	@echo "Resetting Docker DB..."
	@sleep 2
	docker exec -i pms-db psql -U $(DB_USER) -d $(DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	cat migrations/*.up.sql | docker exec -i pms-db psql -U $(DB_USER) -d $(DB_NAME)
	cat scripts/seed_data.sql | docker exec -i pms-db psql -U $(DB_USER) -d $(DB_NAME)
