# PMS Backend Core

Backend for a Hotel Management System (Property Management System) designed with a focus on scalability, data consistency, and security.

The system utilizes a "Compute on Read" architecture for calculating prices and availability, avoiding inventory desynchronization. It is built following Clean Architecture and Multi-tenancy principles.

## Technologies

- Language: Go (Golang) 1.23+
- Database: PostgreSQL 15+
- Web Framework: Echo v4
- SQL Driver: pgx/v5
- Infrastructure: Docker & Docker Compose
- Automation: GNU Make

## Key Features

- Clean Architecture: Strict separation of layers (Handler, Usecase, Repository, Entity).
- Multi-Tenancy: Support for multiple hotels and owners within the same instance.
- Dynamic Pricing Engine: Real-time rate calculation based on rules and priorities.
- Transactional Availability: Overbooking prevention through ACID transactions and database-level locking.
- Advanced Security: JWT Authentication with unique Salt per user (allows immediate session revocation).
- Audit: Automatic tracking of creation and updates (created_at, updated_at) and logical deletion (Soft Delete).

## Prerequisites

To run this project, you need to have the following installed:

1. Go 1.23 or higher
2. Docker and Docker Compose
3. Make (usually included in Linux/Mac, or via Chocolatey/Scoop on Windows)
4. PostgreSQL Client (psql) - Optional but recommended for debugging

## Configuration

Create a .env file in the project root by copying the following content:

```bash
DB_USER=postgres
DB_PASSWORD=postgres
DB_HOST=localhost
DB_PORT=5432

DB_NAME=hotel_pms_db
DB_TEST_NAME=hotel_pms_test

PORT=8081
```

# JWT_SECRET is not required as we use dynamic salts per user in the DB

## Running the Project

The project includes a Makefile to simplify all common tasks.

### Option A: Running with Docker (Recommended)

Starts the database and the API in isolated containers. The API will be available on port 4000.

1. Start services:

  ```bash
  make docker-up
  ```

2. Initialize database (Only the first time or to reset):
   Creates the schema, applies migrations, and loads seed data.

  ```bash
  make docker-db-reset
  ```

3. View logs:

  ```bash
  make docker-logs
  ```

4. Stop services:

  ```bash
  make docker-down
  ```

### Option B: Local Execution (Development)

Requires a locally running PostgreSQL instance on port 5432.

1. Prepare local database:

  ```bash
  make db-reset
  ```

2. Start the server (Hot reload not included, restart manually):

  ```bash
  make run
  ```

The server will listen on the port defined in .env (default 8081).

## Testing

The project features an integration test suite (End-to-End) that validates complete business flows against a real test database.

- Run all tests:

  ```bash
  make test-all
  ```

- Run only unit tests (without DB):

  ```bash
  make test-unit
  ```

- Run only the lifecycle test (Full flow):

  ```bash
  make test-lifecycle
  ```

## Project Structure

```text
pms-core/
├── cmd/
│   └── /api          # Entry point (main.go)
├── internal/
│   ├── bootstrap/    # Configuration and dependency injection
│   ├── entity/       # Domain models and errors
│   ├── handler/      # HTTP Controllers (JSON Input/Output)
│   ├── usecase/      # Pure business logic
│   ├── repository/   # Data access (SQL queries)
│   └── security/     # Authentication middlewares and utilities
├── migrations/       # SQL scripts for DB structure
├── scripts/          # Seed data
└── tests/            # E2E integration tests
```
