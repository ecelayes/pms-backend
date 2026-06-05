# ADR-0001: Hexagonal Architecture (Ports & Adapters)

**Status**: Accepted
**Date**: 2026-06-05
**Deciders**: Engineering team

## Context

Early versions of the codebase followed a "Clean Architecture" pattern with `handler/` / `usecase/` / `repository/` directories. As the system grew, this led to:

- **Leaky abstractions**: HTTP DTOs leaking into business logic; SQL errors leaking into handlers
- **Testability friction**: Application layer could not be tested without a database
- **Unclear ownership**: Where does a `Pinger` for health checks live? Who owns the JWT validator?
- **Bounded-context drift**: All contexts shared the same `internal/entity/`, causing the `entity` package to bloat

## Decision

We adopt **Hexagonal Architecture** (a.k.a. Ports & Adapters):

- **Domain** owns the truth: entities, value objects, repository **interfaces** (ports), domain events
- **Application** owns use cases: orchestration, depends on **domain ports only**
- **Adapter** owns IO: HTTP handlers, Postgres repositories, Redis clients, third-party integrations

```
internal/<context>/
├── domain/         # NO external imports; pure Go
├── application/    # Imports domain + shared/domain only
└── adapter/
    ├── http/       # Echo handlers, HTTP-specific validation
    └── persistence/ # Postgres, Redis, etc.
```

### Bounded contexts

- `iam` — Identity & Access Management
- `catalog` — Properties, units, amenities
- `pricing` — Rate plans, pricing rules
- `availability` — Per-day inventory
- `booking` — Reservations, lifecycle, events

Each bounded context ships its own `domain`, `application`, and `adapter(s)`. Cross-cutting concerns live in `internal/shared/`.

### Ports vs adapters

A **port** is a Go interface that the domain or application layer defines, expressing what it needs. An **adapter** is a concrete implementation (e.g., `pgx.UserRepository`) that satisfies the port.

```go
// internal/iam/domain/repository.go (port)
type UserRepository interface {
    FindByEmail(ctx context.Context, email string) (*User, error)
}

// internal/iam/adapter/persistence/postgres/user_repository.go (adapter)
type UserRepository struct { db *pgxpool.Pool }
func (r *UserRepository) FindByEmail(...) (*User, error) { ... }
```

## Consequences

### Positive
- **Testability**: Application tests run without a database; mocks satisfy ports
- **Replaceability**: Swapping Postgres for another store means a new adapter, not changes to domain
- **Bounded autonomy**: Each context can evolve independently; no global `entity` package
- **Clear contracts**: Ports document what the application needs; reviewable in isolation

### Negative
- **More files**: 3+ files per concept (port + impl + tests) instead of one
- **Indirection**: New developers must learn to navigate the layering
- **Verbose mocking**: Port interfaces can have many methods, leading to large mocks
- **Refactor cost**: Moving a concept between layers is non-trivial (but rare)

## Mitigations

- We rely on `mockery` (or hand-rolled mocks when small) to keep test code DRY
- The `AGENTS.md` documents the layer rules and gives a 4-phase checklist for new features
- A `golangci-lint` rule set is in place to catch domain-layer imports of external packages

## Alternatives considered

- **Clean Architecture (handler/usecase/repository)**: rejected — it does not enforce a single direction of dependency at the language level
- **Microservices**: rejected — overhead too high for a single-tenant PMS, and the team is small
- **Modular monolith with Go packages only**: rejected — too easy to break the rules without explicit `domain`/`application`/`adapter` boundaries
