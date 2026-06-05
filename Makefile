# Database utilities
# Usage: make db-setup db-create db-migrate-up db-migrate-down db-status

.PHONY: db-setup
db-setup:
	go run cmd/migrate/main.go setup

.PHONY: db-create
db-create:
	go run cmd/migrate/main.go create-db

.PHONY: db-migrate-up
db-migrate-up:
	go run cmd/migrate/main.go migrate-up

.PHONY: db-migrate-down
db-migrate-down:
	go run cmd/migrate/main.go migrate-down

.PHONY: db-status
db-status:
	go run cmd/migrate/main.go status

.PHONY: db-reset
db-reset:
	go run cmd/migrate/main.go migrate-down
	go run cmd/migrate/main.go migrate-up

# Test with database
.PHONY: test-db
test-db:
	@echo "Run: TEST_DATABASE_URL=\$$(grep DB_TEST_NAME .env | cut -d'=' -f2 | xargs echo -n)://\$$(grep DB_USER .env | cut -d'=' -f2 | xargs):\$$(grep DB_PASSWORD .env | cut -d'=' -f2 | xargs)@localhost:5432/\$$(grep DB_TEST_NAME .env | cut -d'=' -f2 | xargs)?sslmode=disable go test ./tests/..."

# Development server
.PHONY: dev
dev:
	docker compose up -d db redis
	make db-setup
	go run cmd/api/main.go

# Install golang-migrate tool
.PHONY: install-migrate
install-migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Clean docker volumes (WARNING: destroys data)
.PHONY: db-clean
db-clean:
	docker compose down -v

# Run integration tests against the real Redis + Postgres
# Requires the env vars: TEST_DATABASE_URL, TEST_REDIS_ADDR
.PHONY: test-integration
test-integration:
	@if [ ! -f .env ]; then echo "Missing .env"; exit 1; fi
	@set -a && . ./.env && set +a; \
		export TEST_DATABASE_URL=postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_TEST_NAME?sslmode=disable; \
		export TEST_REDIS_ADDR=$${TEST_REDIS_ADDR:-localhost:6379}; \
		go test -count=1 -timeout=120s ./tests/...

# Run unit tests + integration tests in sequence
.PHONY: test-all
test-all: test
	@set -a && . ./.env && set +a; \
		export TEST_DATABASE_URL=postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_TEST_NAME?sslmode=disable; \
		export TEST_REDIS_ADDR=$${TEST_REDIS_ADDR:-localhost:6379}; \
		go test -count=1 -timeout=120s ./tests/...


# ───────────────────────────────────────────────────────────
# Linting
# ───────────────────────────────────────────────────────────
.PHONY: lint
lint:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not installed. Run: make lint-install"; \
		exit 1; \
	fi
	golangci-lint run ./...

.PHONY: lint-install
lint-install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: lint-fix
lint-fix:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not installed. Run: make lint-install"; \
		exit 1; \
	fi
	golangci-lint run --fix ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: install-hooks
install-hooks:
	git config core.hooksPath .githooks
	@echo "Pre-commit hooks installed. They run on every commit."

.PHONY: uninstall-hooks
uninstall-hooks:
	git config --unset core.hooksPath
	@echo "Pre-commit hooks removed."
