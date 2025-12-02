ifneq (,$(wildcard ./.env))
    include .env
    export
endif

APP_NAME=pms-backend
DB_DSN=$(DATABASE_URL)

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
	psql "$(DB_DSN)" -f migrations/001_initial_schema.up.sql

db-seed: db-check
	@echo "Seeding data"
	psql "$(DB_DSN)" -f scripts/seed_data.sql

db-reset: db-check
	@echo "Resetting Database State"
	psql "$(DB_DSN)" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@make db-up
	@make db-seed
	@echo "Database ready for testing!"
