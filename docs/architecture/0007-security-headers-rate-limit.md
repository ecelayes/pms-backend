# ADR-0007: Security Headers and Per-IP Rate Limiting

**Status**: Accepted
**Date**: 2026-06-05

## Context

Public-facing APIs are routinely probed for common vulnerabilities. The OWASP Secure Headers project lists a baseline of HTTP response headers that mitigate common attacks:

- `X-Content-Type-Options: nosniff` — prevents MIME-sniffing
- `X-Frame-Options: DENY` — prevents clickjacking
- `Referrer-Policy: strict-origin-when-cross-origin` — limits referrer leakage
- `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'` — strict CSP for APIs
- `Strict-Transport-Security` — HSTS, but **only over TLS** (sending it over plain HTTP leaks the policy)

Separately, auth endpoints are a brute-force target. We need to throttle per-IP without affecting the rest of the API.

## Decision

### 1. Security headers

We add a global middleware in `internal/shared/adapter/http/security_headers.go`:

```go
func SecurityHeaders() echo.MiddlewareFunc { return SecurityHeadersWithConfig(DefaultSecurityHeadersConfig) }
```

Defaults:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `X-XSS-Protection: 0` (disables the buggy old filter; modern browsers rely on CSP)
- `Cross-Origin-Opener-Policy: same-origin`
- `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'`
- `Strict-Transport-Security: max-age=63072000; includeSubDomains` — **only on TLS**

The middleware uses a `setIfMissing` helper that **does not override** headers already set by the application (e.g., a custom CSP for a documentation route).

Wired in `internal/bootstrap/app.go` **after CORS** (so CORS headers are not stripped).

### 2. Rate limiting

We add a per-IP token-bucket limiter using `golang.org/x/time/rate`:

```go
type IPRateLimiter struct {
    rate     rate.Limit    // tokens/sec
    burst    int           // bucket size
    ttl      time.Duration // bucket eviction
    mu       sync.Mutex
    buckets  map[string]*ipBucket
}
```

Defaults: 5 req/s, burst 10, TTL 10 minutes. A background goroutine evicts stale buckets every minute.

Wired in `internal/iam/module.go` on the `/auth` group **only**:

```go
authGroup := group.Group("/auth", authLimiter.Middleware())
```

### 3. Tests

- `security_headers_test.go`: 5 tests, 100% coverage, including HSTS-not-on-HTTP, setIfMissing behavior
- `rate_limiter_test.go`: 3 unit tests + 2 race tests, 100% coverage

## Consequences

### Positive
- **Defense in depth**: a single forgotten header in a future handler is fine; the middleware sets it
- **Brute-force protection**: 5 req/s means a credential-stuffing attack on `/auth/login` is throttled
- **Observability**: every 429 response logs the IP and route at `warn` level
- **Testability**: the limiter's `Allow()` method is pure; tests do not need the network

### Negative
- **False positives** on shared NATs: a corporate office behind one IP would be limited. Mitigation: 5 req/s is generous for a login flow
- **Memory**: the bucket map grows with unique IPs. The GC loop bounds it (TTL = cleanup interval)

### Mitigations
- The limiter's TTL is configurable via env var in a follow-up
- Behind a trusted proxy, operators can swap `c.RealIP()` for `X-Forwarded-For`; we document this in deployment notes
- Tests for the limiter include a 50-goroutine race test

## Alternatives considered

- **Echo's built-in `middleware.RateLimiter`**: rejected — it uses a single token bucket, not per-IP
- **Nginx/Envoy rate limiting**: rejected — we want it at the application for now, easier to test
- **Cloudflare / WAF**: planned for production, but we want defense in depth in the app too
