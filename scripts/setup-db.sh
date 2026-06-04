#!/bin/bash
# Setup database and run migrations
# Usage: ./scripts/setup-db.sh

set -e

# Load .env
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

MIGRATE_BIN="${HOME}/go/bin/migrate"
if [ ! -f "$MIGRATE_BIN" ]; then
    echo "=== Installing golang-migrate ==="
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}?sslmode=disable"
TEST_DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_TEST_NAME}?sslmode=disable"

echo ""
echo "=== Creating test database ==="
# Try psql first, fallback to createdb
if command -v psql &> /dev/null; then
    psql "${DB_URL}" -c "CREATE DATABASE ${DB_TEST_NAME}" 2>/dev/null || echo "DB ${DB_TEST_NAME} already exists"
elif command -v createdb &> /dev/null; then
    createdb -h localhost -U "${DB_USER}" "${DB_TEST_NAME}" 2>/dev/null || echo "DB ${DB_TEST_NAME} already exists"
else
    echo "WARNING: psql/createdb not found. Ensure ${DB_TEST_NAME} exists manually."
fi

echo ""
echo "=== Running migrations ==="
"$MIGRATE_BIN" -path db/migrations -database "${DB_URL}" up
"$MIGRATE_BIN" -path db/migrations -database "${TEST_DB_URL}" up

echo ""
echo "=== Migration status ==="
"$MIGRATE_BIN" -path db/migrations -database "${DB_URL}" version
echo ""
echo "Done! Test with:"
echo "  TEST_DATABASE_URL=\"${TEST_DB_URL}\" go test ./tests/..."
