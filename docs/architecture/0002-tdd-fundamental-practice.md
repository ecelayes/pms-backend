# ADR-0002: Tests as the Fundamental Practice (TDD)

**Status**: Accepted
**Date**: 2026-06-05
**Deciders**: Engineering team

## Context

Historically, tests were treated as a verification step, written *after* code. This led to:

- **Coverage theater**: Tests that exercise the happy path but miss edge cases
- **Brittle tests**: Tests coupled to implementation, not to spec
- **Late feedback**: Bugs found at integration, not unit, level
- **Slow velocity**: Refactors required rewriting large test files

## Decision

We adopt **Test-Driven Development (TDD)** as a fundamental practice. Tests are the **specification** of behavior, not the verification of code. The cycle is:

1. **Red** — write a failing test that describes a behavior
2. **Green** — write the minimum code that makes the test pass
3. **Refactor** — improve the code while keeping the test green
4. **Commit** — small, atomic commits per cycle

### Coverage gates

- **Application layer**: 100% coverage. Domain logic is pure and must be exhaustively tested.
- **Domain layer**: 100% coverage. VOs and entities have explicit sentinels for every failure path.
- **Adapter layer (HTTP)**: focus on validation, binding, and error mapping; the heavy logic is in application.
- **Adapter layer (persistence)**: integration tests run against a real Postgres; unit tests use sqlmock.

CI enforces a **80% coverage gate** for the project. PRs that drop coverage are rejected.

### Test taxonomy

- **Unit tests**: no DB, no Redis, no Echo; pure functions in the same package (`_test.go`).
- **Integration tests**: real Postgres + real Redis; run via `make test-integration`. The CI spins up ephemeral services.
- **Race tests**: critical paths (rate limiter, inventory) run with `go test -race`.
- **Race conditions**: tests for shared state (e.g., concurrent reservation creation) live in `*_race_test.go`.

### Mocking strategy

- **Ports** (interfaces) in domain/application are mocked. We use hand-rolled mocks for small interfaces (< 5 methods); `mockery` for larger ones.
- **Adapters** are not mocked in unit tests; their behavior is verified by integration tests.

## Consequences

### Positive
- **Confidence to refactor**: 100% application coverage means we can change internals freely
- **Living spec**: The test file is the documentation of what the code does
- **Faster feedback**: Bugs caught at unit-test time, not at integration
- **Easier code review**: Tests reveal intent before code does

### Negative
- **Slower initial development**: Writing tests first takes longer than typing code first
- **Test maintenance**: Refactors may require updating many tests, but only if tests were coupled to implementation
- **Coverage pressure**: A 100% target can push developers to write tests for trivial getters

### Mitigations
- We use `// coverage-ignore` only with explicit justification; the linter warns on un-tagged coverage holes
- We review tests in PRs, not just code
- We do not write tests for trivial accessors (e.g., `func (p *Password) IsEmpty() bool { return *p == "" }`)

## Alternatives considered

- **BDD with Gherkin**: rejected — adds a translation layer between spec and code
- **Property-based testing**: rejected as a default; we use `testing/quick` only for parsers
- **Mutation testing**: rejected for now; high false-positive rate with mocks
