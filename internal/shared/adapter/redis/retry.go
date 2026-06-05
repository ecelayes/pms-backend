package redis

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"time"
)

// ErrNonRetryable is wrapped by callers to signal that withRetry must not
// retry the operation even if attempts remain.
var ErrNonRetryable = errors.New("non-retryable")

// retryConfig describes a retry policy.
//
//	MaxAttempts:   total attempts including the first call (>= 1)
//	InitialBackoff: wait before the second attempt
//	Multiplier:   factor applied to backoff on each failed attempt
//	MaxBackoff:   cap for a single backoff interval
//	Jitter:       random factor in [0, 1) added to each backoff
type retryConfig struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	Multiplier     float64
	MaxBackoff     time.Duration
	Jitter         float64
}

func defaultRetryConfig() retryConfig {
	return retryConfig{
		MaxAttempts:    3,
		InitialBackoff: 50 * time.Millisecond,
		Multiplier:     2.0,
		MaxBackoff:     500 * time.Millisecond,
		Jitter:         0.2,
	}
}

func (c retryConfig) backoffFor(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	d := float64(c.InitialBackoff)
	for i := 1; i < attempt; i++ {
		d *= c.Multiplier
		if time.Duration(d) >= c.MaxBackoff {
			return c.MaxBackoff
		}
	}
	if time.Duration(d) > c.MaxBackoff {
		return c.MaxBackoff
	}
	if c.Jitter > 0 {
		d += d * c.Jitter * rand.Float64()
	}
	return time.Duration(d)
}

// isRetryable returns true if the error should trigger a retry.
// Network errors, timeouts, and Redis "LOADING"/"BUSY" responses are retryable.
// Application errors (validation, not-found) are not.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrNonRetryable) {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "connection reset"),
		strings.Contains(msg, "broken pipe"),
		strings.Contains(msg, "i/o timeout"),
		strings.Contains(msg, "EOF"),
		strings.Contains(msg, "LOADING"),
		strings.Contains(msg, "BUSY"),
		strings.Contains(msg, "READONLY"),
		strings.Contains(msg, "redis: "):
		return true
	}
	return false
}

// withRetry executes op under the given retry policy.
//
// It retries only when isRetryable returns true. The first call is attempt 1.
// Backoff uses exponential growth with jitter, capped at MaxBackoff.
func withRetry[T any](ctx context.Context, cfg retryConfig, op func(ctx context.Context) (T, error)) (T, error) {
	var zero T
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = 1
	}
	var lastErr error
	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return zero, err
		}
		v, err := op(ctx)
		if err == nil {
			return v, nil
		}
		lastErr = err
		if !isRetryable(err) {
			return zero, err
		}
		if attempt == cfg.MaxAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(cfg.backoffFor(attempt)):
		}
	}
	return zero, lastErr
}
