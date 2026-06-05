# ADR-0005: Retry with Exponential Backoff + Jitter (generic)

**Status**: Accepted
**Date**: 2026-06-05

## Context

Redis calls can fail transiently: a network blip, a failover, a brief CPU spike on the server. The naive response is to fail the request and have the client retry — but that pushes complexity up the stack.

We initially had no retry logic. Symptoms:
- Booking creation would fail 1 in 50 times under load
- The DLQ would fill up with messages that just needed a second try
- We had no way to distinguish transient from permanent failures

## Decision

We implement a **generic retry helper** in `internal/shared/adapter/redis/retry.go`:

```go
func withRetry[T any](ctx context.Context, cfg RetryConfig, fn func(context.Context) (T, error)) (T, error)
```

### Defaults

| Setting      | Value   |
|--------------|---------|
| Max attempts | 3       |
| Initial wait | 50ms    |
| Multiplier   | 2x      |
| Max wait     | 500ms   |
| Jitter       | ±20%    |

### Which errors are retryable

Only **network-level** and **Redis-protocol-level** errors retry:
- `context.DeadlineExceeded` (only if the deadline is NOT the outer request's)
- Connection refused / reset
- `redis.ErrClosed` (during failover)
- Timeouts

Application-level errors (e.g., `domain.ErrNotFound`, `pgx.ErrNoRows`) **never retry**. They are deterministic.

### Wrapping non-retryable errors

We define a sentinel `ErrNonRetryable` that callers wrap to opt out of retry:

```go
return "", fmt.Errorf("permanent: %w: %w", ErrNonRetryable, err)
```

### Wiring

`EventPublisher.publish` wraps its `p.client.Publish` call in `withRetry`. The call site is unchanged from a caller's perspective:

```go
if err := p.client.Publish(...).Err(); err != nil { return err }
```

becomes:

```go
_, err := withRetry(ctx, defaultRetryConfig(), func(ctx context.Context) (int64, error) {
    return p.client.Publish(ctx, ...).Result()
})
return err
```

## Consequences

### Positive
- **Resilience**: bookings succeed even when Redis hiccups
- **Tunable**: configuration lives in one place
- **Type-safe**: generics avoid `interface{}` for the success value
- **Testable**: we have unit tests for success-after-transient, max-attempts, context-cancel, and non-retryable paths

### Negative
- **Latency**: under failure, a 3-attempt retry adds up to 50 + 100 + 200 = 350ms (with jitter)
- **Cost**: each retry is a real round-trip; we set tight bounds

### Mitigations

- The outer request context **always** wins. If the user cancels, retry stops immediately
- Jitter prevents thundering-herd when many clients retry at once
- We log each retry at `warn` level with the attempt number and error (no PII)

## Alternatives considered

- **Hystrix-style circuit breaker**: rejected for now — Redis rarely fails long enough to trip the breaker; the retry handles the short cases
- **Libraries like `cenkalti/backoff`**: rejected — we wanted zero new dependencies and full control over the isRetryable filter
- **No retry, just fail fast**: rejected — we measured 1-2% flaky failure rate in dev
