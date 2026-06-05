# PMS Backend Core

Backend for a Property Management System (PMS) designed with a focus on scalability, data consistency, and security.

The system uses a **"Compute on Read"** architecture for prices and availability, avoiding inventory desynchronization. It is built following **Hexagonal Architecture** (ports & adapters) and **Domain-Driven Design** principles, with strict TDD.

## Technologies

- Language: Go (Golang) 1.24+
- Database: PostgreSQL 16+
- Cache / Streams / DLQ: Redis 7+
- Web Framework: Echo v4
- SQL Driver: pgx/v5
- Logging: Uber Zap (structured)
- Validation: TDD-first, custom validators in `internal/shared/adapter/http`
- Linting: golangci-lint v1.64.8
- Automation: GNU Make + GitHub Actions

## Key Features

- **Hexagonal Architecture**: Bounded contexts (`iam`, `catalog`, `pricing`, `availability`, `booking`) with domain ports + adapters
- **Multi-tenancy**: Organizations own properties, members have roles
- **Dynamic Pricing Engine**: Real-time rate calculation with rate plans
- **Transactional Availability**: Overbooking prevention via row-level locks + count check
- **Advanced Security**:
  - Bcrypt password hashing
  - Per-user salt in JWT (immediate session revocation)
  - Argon2-style header validation
  - Rate limiting on auth endpoints (5 req/s burst 10)
  - Security headers (CSP, HSTS-on-TLS, X-Frame-Options, etc.)
  - Password strength validation (length, variety, common-password blocklist)
- **Observability**:
  - Structured logging with request IDs propagated through async goroutines
  - Liveness (`/live`) and Readiness (`/ready`) probes with `Pinger` interface
- **Resilience**:
  - Retry with exponential backoff + jitter for Redis ops
  - Dead-letter queue for failed stream messages
  - Graceful shutdown with timeouts
- **CQRS-style events**: Reservation lifecycle emits `reservation.created`, `.confirmed`, `.cancelled` via Redis Streams

## Bounded Contexts

```
internal/
├── iam/          # Identity & Access Management
├── catalog/      # Properties, UnitTypes, Units, Amenities, Services
├── pricing/      # Pricing rules + rate plans
├── availability/ # Per-day availability (Redis)
├── booking/      # Reservation lifecycle
└── shared/       # Cross-cutting: errors, request_id, validation, Redis ports
```

## Prerequisites

- Go 1.24+
- PostgreSQL 16 (running on `localhost:5432`)
- Redis 7 (running on `localhost:6379`)
- Make
- golangci-lint (optional, `make lint-install`)

## Configuration

Copy `.env.example` to `.env` and adjust:

```bash
DB_USER=postgres
DB_PASSWORD=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=pms_db
DB_TEST_NAME=pms_db_test
PORT=8080
TEST_REDIS_ADDR=localhost:6379

# Server timeouts (Go duration syntax or seconds)
SERVER_READ_HEADER_TIMEOUT=10s
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
SERVER_IDLE_TIMEOUT=120s
SERVER_SHUTDOWN_TIMEOUT=15s

# CORS
CORS_ALLOWED_ORIGINS=*
```

JWT secret is **not** required: we use per-user salts stored in the DB, allowing immediate token revocation.

## Running

### Quick start (local)

```bash
# 1. Create database + run migrations
make db-reset

# 2. Run the API
make run
```

The server listens on `$PORT` (default 8080).

### Available make targets

```bash
make help              # List all targets
make run               # Run the API
make db-create         # Create the database
make db-migrate-up     # Apply migrations
make db-status         # Show migration status
make db-reset          # Drop, recreate, migrate, seed

make test              # Unit tests only
make test-integration  # Integration tests (needs DB + Redis)
make test-all          # All tests
make test-coverage     # Unit tests with coverage report

make lint              # Run golangci-lint
make lint-fix          # Auto-fix lint issues
make lint-install      # Install golangci-lint
make vet               # Run go vet

make build             # Build binaries to dist/
make docker-build      # Build Docker image
make install-hooks     # Install pre-commit hooks
```

## API

OpenAPI 3.0 spec: [`docs/api/openapi.yaml`](docs/api/openapi.yaml).

Health checks:
- `GET /live` — liveness
- `GET /ready` — readiness (checks Postgres + Redis)
- `GET /health` — legacy alias

API routes are under `/api/v1/...`; admin DLQ routes under `/admin/dlq/...`.

Authentication: `Authorization: Bearer <jwt>` on all `/api/v1/*` except `/api/v1/auth/*` and `/api/v1/reservations` (guest can self-register).

## Testing

We follow **TDD** strictly: tests are written first as the specification. Coverage targets: **80%+ per package**, with 100% on `application` and `domain` layers.

```bash
make test           # Unit tests with -race
make test-coverage  # With coverage report
```

Integration tests run against the real Postgres + Redis:

```bash
make db-reset
make test-integration
```

## Project Structure

```
pms-backend/
├── cmd/
│   ├── api/                # HTTP server entry point
│   └── migrate/            # Database migration utility
├── internal/
│   ├── bootstrap/          # DI, health, security headers, rate limiter
│   ├── iam/                # Bounded context
│   │   ├── domain/         # Entities + VOs + errors (no deps)
│   │   ├── application/    # Use cases (depends on domain only)
│   │   └── adapter/        # HTTP handlers + Postgres repo
│   ├── catalog/
│   ├── pricing/
│   ├── availability/
│   ├── booking/
│   └── shared/             # Cross-cutting
├── db/migrations/          # SQL migrations (consolidated)
├── docs/
│   ├── api/openapi.yaml    # OpenAPI 3.0 spec
│   └── architecture/       # ADRs and standards
├── tests/                  # E2E integration tests
├── .github/                # CI/CD workflows + PR template
├── .githooks/              # Pre-commit hooks
├── Makefile
├── Dockerfile
└── go.mod
```

## CI/CD

GitHub Actions workflows:

- **CI** (`.github/workflows/ci.yml`): unit tests (with `-race` + 80% coverage gate), lint, integration tests against ephemeral Postgres + Redis
- **Release** (`.github/workflows/release.yml`): multi-platform binaries on `v*` tag push

Dependabot updates Go modules and GitHub Actions weekly.

## Documentation

- [`.agents/docs/IMPROVEMENT_PLAN.md`](.agents/docs/IMPROVEMENT_PLAN.md) — 30 improvement items across 6 stages
- [`docs/architecture/`](docs/architecture/) — Architecture Decision Records (ADRs) and standards
- [`docs/api/openapi.yaml`](docs/api/openapi.yaml) — OpenAPI 3.0 spec
- [`AGENTS.md`](AGENTS.md) — Agent harness engineering rules

## Security

- Passwords: Bcrypt, never stored or logged
- JWT: per-user salt in `users.salt` column; signature is verified on every request
- Rate limiting: 5 req/s burst 10 on `/api/v1/auth/*`
- Security headers: CSP, X-Frame-Options, HSTS-on-TLS, etc.
- Input validation: email format, UUIDs, required fields
- Password strength: min 8 chars, upper + (lower or digit), not in common-password list

To report a security issue, see [SECURITY.md](SECURITY.md) (if present).

## License

[Add license here]
