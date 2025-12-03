ifneq (,$(wildcard ./.env))
    include .env
    export
endif

APP_NAME=pms-backend
DB_DSN=$(DATABASE_URL)
DB_TEST_DSN=$(TEST_DATABASE_URL)

.PHONY: help run build test clean db-up db-seed db-reset

help:
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run:
	@echo "Starting $(APP_NAME)"
	go run cmd/api/main.go

build:
	@echo "Building binary"
	go build -o bin/$(APP_NAME) cmd/api/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/
	go clean

# ==============================================================================
# Database Helpers (Using native psql)
# ==============================================================================

db-check:
	@if [ -z "$(DB_DSN)" ]; then echo "Error: DATABASE_URL is not set in .env"; exit 1; fi

db-up: db-check
	@echo "Applying schema"
	@cat migrations/*.up.sql | psql "$(DB_DSN)"

db-seed: db-check
	@echo "Seeding data"
	psql "$(DB_DSN)" -f scripts/seed_data.sql

db-reset: db-check
	@echo "Resetting Database State"
	psql "$(DB_DSN)" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@make db-up
	@make db-seed
	@echo "Database ready for testing!"

# ==============================================================================
# Testing Helpers
# ==============================================================================

test-prepare:
	@echo "Preparing Test Database"
	psql "$(DB_TEST_DSN)" -c "DROP DATABASE IF EXISTS hotel_pms_test;"
	psql "$(DB_TEST_DSN)" -c "CREATE DATABASE hotel_pms_test;"
	
	@make db-up DATABASE_URL="$(DB_TEST_DSN)"

test-e2e: test-prepare
	@echo "Running E2E Tests"
	export TEST_DATABASE_URL="$(DB_TEST_DSN)"; go test -v ./tests/...
