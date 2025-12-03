# VAR
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

APP_NAME=pms-backend

DB_DSN ?= $(DATABASE_URL)

TEST_DB_DSN ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_TEST_NAME)?sslmode=disable

DB_ADMIN_DSN := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/postgres?sslmode=disable

.PHONY: help run build test clean db-up db-seed db-reset docker-up docker-down docker-db-reset

help:
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# APP
run:
	@echo "Starting $(APP_NAME)..."
	go run cmd/api/main.go

build:
	@echo "Building binary..."
	go build -o bin/$(APP_NAME) cmd/api/main.go

clean:
	rm -rf bin/
	go clean

# DB
db-check:
	@if [ -z "$(DB_DSN)" ]; then echo "Error: DATABASE_URL is not set in .env"; exit 1; fi

db-up: db-check
	@echo "Applying schema..."
	@cat migrations/*.up.sql | psql "$(DB_DSN)"

db-seed: db-check
	@echo "Seeding data..."
	@cat scripts/seed_data.sql | psql "$(DB_DSN)"

db-reset: db-check
	@echo "Resetting Database State..."
	psql "$(DB_DSN)" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@make db-up
	@make db-seed
	@echo "Database ready!"

# TEST
test-prepare:
	@echo "Preparing Test Database: $(DB_TEST_NAME)..."
	@psql "$(DB_ADMIN_DSN)" -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$(DB_TEST_NAME)' AND pid <> pg_backend_pid();" || true
	
	@psql "$(DB_ADMIN_DSN)" -c "DROP DATABASE IF EXISTS $(DB_TEST_NAME);"
	@psql "$(DB_ADMIN_DSN)" -c "CREATE DATABASE $(DB_TEST_NAME);"
	
	@make db-up DATABASE_URL="$(TEST_DB_DSN)"

test-e2e: test-prepare
	@echo "Running E2E Tests..."
	export TEST_DATABASE_URL="$(TEST_DB_DSN)"; go test -v ./tests/...

# DOCKER
docker-up:
	docker-compose up -d --build
	@echo "Docker running on http://localhost:4000"

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f api

docker-db-reset:
	@echo "Resetting Docker Database..."
	@sleep 2
	docker exec -i pms-db psql -U $(DB_USER) -d $(DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	cat migrations/*.up.sql | docker exec -i pms-db psql -U $(DB_USER) -d $(DB_NAME)
	cat scripts/seed_data.sql | docker exec -i pms-db psql -U $(DB_USER) -d $(DB_NAME)
	@echo "Docker Database Ready!"
