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
