# ADR-0003: Redis as a Hexagonal Adapter (No miniredis in unit tests)

**Status**: Accepted
**Date**: 2026-06-05
**Deciders**: Engineering team

## Context

The codebase used Redis Streams for reservation events. Early on, the use cases depended on `*redis.Client` directly:

```go
// BEFORE
type AvailabilityService struct {
    redis *redis.Client // concrete dependency
}
```

Problems:
- **Tests needed a real Redis** OR `miniredis` (an in-process Redis stub)
- **Adapter details leaked into application**: stream names like `reservations.events` were hard-coded across packages
- **Coupling**: a Redis library upgrade required changes in the application layer

We also tried `miniredis` in unit tests. The result was:
- **Subtle behavior drift**: miniredis does not implement Streams identically (DLQ, consumer groups differ)
- **Maintenance burden**: miniredis was always a version behind `redis-server`
- **False sense of coverage**: passing miniredis tests sometimes failed against real Redis

## Decision

### 1. Redis moves to a true hexagonal adapter

- **Domain ports** in `internal/shared/domain/ports.go` declare `EventPublisher`, `DLQStore`, `Cache`, etc.
- **Redis adapter** in `internal/shared/adapter/redis/` implements those ports
- **Stream names, group names, and constants** are implementation details of the adapter — they never appear in the application layer

```go
// internal/shared/domain/ports.go (port)
type EventPublisher interface {
    Publish(ctx context.Context, stream string, payload []byte) error
}

// internal/shared/adapter/redis/event_publisher.go (adapter)
const StreamReservations = "reservations.events"
type EventPublisher struct{ client *redis.Client }
func (p *EventPublisher) Publish(...) error { ... }
```

### 2. No `miniredis` in unit tests

- **Unit tests** of the application layer mock the **port**, not the adapter
- **Integration tests** run against a real Redis 7 instance, started by the developer or CI
- We added a `tests/integration/` test that exercises Streams + consumer groups + DLQ end-to-end

### 3. Adapter-level unit tests use the real client (where possible)

For adapter behavior that is not easily testable via integration (e.g., retry on transient network errors), we use a fake client implementing the same surface, OR run with `TEST_REDIS_ADDR` set.

## Consequences

### Positive
- **Application tests are fast** (microseconds per test) — no network, no miniredis startup
- **Test fidelity**: integration tests catch what miniredis hides (e.g., XTRIM edge cases)
- **Refactor-friendly**: swapping Redis for another stream engine (Kafka, NATS) is a single new adapter
- **One source of truth** for stream names — change the constant once in the adapter

### Negative
- **Two test layers**: developers must remember to run `make test-integration` before pushing
- **CI infrastructure**: integration tests need a Redis service; GitHub Actions provides one
- **No fake Redis in dev**: developers must have Redis running locally (documented in README)

### Mitigations
- `make test-integration` and `make test-all` are explicit and well-documented
- The `Makefile` target prints a clear message if `TEST_REDIS_ADDR` is unreachable
- CI fails fast if Redis is not available

## Alternatives considered

- **Keep miniredis**: rejected — coverage was high but correctness was questionable
- **Use testcontainers-go**: rejected for unit tests (too slow); we use it for a separate "smoke" suite in CI
- **Drop Redis Streams entirely**: rejected — we need a real-time event channel for booking lifecycle
