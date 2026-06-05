package redis

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

var errTransient = errors.New("redis: connection refused")

type flakyOp struct {
	failures atomic.Int32
	maxFails int32
	result   string
}

func (f *flakyOp) do(ctx context.Context) (string, error) {
	if f.failures.Add(1) <= f.maxFails {
		return "", errTransient
	}
	return f.result, nil
}

func TestRetry_SucceedsAfterTransient(t *testing.T) {
	op := &flakyOp{maxFails: 2, result: "ok"}
	got, err := withRetry(context.Background(), defaultRetryConfig(), func(ctx context.Context) (string, error) {
		return op.do(ctx)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ok" {
		t.Errorf("expected ok, got %q", got)
	}
	if op.failures.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", op.failures.Load())
	}
}

func TestRetry_GivesUpAfterMaxAttempts(t *testing.T) {
	op := &flakyOp{maxFails: 10}
	_, err := withRetry(context.Background(), defaultRetryConfig(), func(ctx context.Context) (string, error) {
		return op.do(ctx)
	})
	if err == nil {
		t.Fatal("expected error after max attempts")
	}
	if !errors.Is(err, errTransient) {
		t.Errorf("expected wrapped transient error, got %v", err)
	}
}

func TestRetry_RespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := withRetry(ctx, defaultRetryConfig(), func(ctx context.Context) (string, error) {
		return "", errTransient
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestRetry_NoRetryOnNonRetryable(t *testing.T) {
	nonRetryable := errors.New("permanent error")
	calls := atomic.Int32{}
	_, err := withRetry(context.Background(), defaultRetryConfig(), func(ctx context.Context) (string, error) {
		calls.Add(1)
		return "", nonRetryable
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls.Load() != 1 {
		t.Errorf("expected 1 call (no retry), got %d", calls.Load())
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	c := defaultRetryConfig()
	if c.MaxAttempts < 2 {
		t.Error("max attempts should be at least 2")
	}
	if c.InitialBackoff <= 0 {
		t.Error("initial backoff should be positive")
	}
	if c.MaxBackoff < c.InitialBackoff {
		t.Error("max backoff should be >= initial backoff")
	}
}

func TestBackoff_IncreasesUpToMax(t *testing.T) {
	c := retryConfig{InitialBackoff: time.Millisecond, MaxBackoff: 100 * time.Millisecond, Multiplier: 2.0}
	d1 := c.backoffFor(1)
	d2 := c.backoffFor(2)
	d3 := c.backoffFor(3)
	if d1 >= d2 || d2 >= d3 {
		t.Errorf("backoff should increase: d1=%v d2=%v d3=%v", d1, d2, d3)
	}
	if d3 > c.MaxBackoff {
		t.Errorf("backoff should cap at max: d3=%v max=%v", d3, c.MaxBackoff)
	}
}


func TestIsRetryable_NetworkErrors(t *testing.T) {
	retryable := []string{
		"redis: connection refused",
		"connection reset by peer",
		"broken pipe",
		"i/o timeout",
		"EOF",
		"LOADING Redis is loading the dataset in memory",
		"BUSY Redis is busy running a script",
		"READONLY You can't write against a read only replica",
		"redis: unexpected error",
	}
	for _, msg := range retryable {
		err := errors.New(msg)
		if !isRetryable(err) {
			t.Errorf("expected retryable: %q", msg)
		}
	}
}

func TestIsRetryable_NonRetryable(t *testing.T) {
	nonRetryable := []error{
		errors.New("validation failed"),
		errors.New("user not found"),
		context.Canceled,
		context.DeadlineExceeded,
	}
	for _, err := range nonRetryable {
		if isRetryable(err) {
			t.Errorf("expected non-retryable: %v", err)
		}
	}
}

func TestIsRetryable_Nil(t *testing.T) {
	if isRetryable(nil) {
		t.Error("nil should not be retryable")
	}
}

func TestIsRetryable_NonRetryableWrapper(t *testing.T) {
	wrapped := fmt.Errorf("op failed: %w", ErrNonRetryable)
	if isRetryable(wrapped) {
		t.Error("ErrNonRetryable should not be retried")
	}
}

func TestBackoff_CapsAtMax(t *testing.T) {
	c := retryConfig{InitialBackoff: 10 * time.Millisecond, MaxBackoff: 50 * time.Millisecond, Multiplier: 10.0}
	if d := c.backoffFor(1); d <= 0 {
		t.Errorf("attempt 1 should be positive, got %v", d)
	}
	if d := c.backoffFor(20); d > c.MaxBackoff {
		t.Errorf("attempt 20 should cap at max %v, got %v", c.MaxBackoff, d)
	}
}


func TestWithRetry_ZeroMaxAttempts(t *testing.T) {
	cfg := retryConfig{MaxAttempts: 0, InitialBackoff: time.Millisecond}
	calls := atomic.Int32{}
	_, err := withRetry(context.Background(), cfg, func(ctx context.Context) (string, error) {
		calls.Add(1)
		return "x", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != 1 {
		t.Errorf("expected 1 call, got %d", calls.Load())
	}
}

func TestWithRetry_LastAttemptDoesNotBackoff(t *testing.T) {
	cfg := retryConfig{MaxAttempts: 2, InitialBackoff: time.Hour}
	op := &flakyOp{maxFails: 5}
	start := time.Now()
	_, err := withRetry(context.Background(), cfg, func(ctx context.Context) (string, error) {
		return op.do(ctx)
	})
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected error")
	}
	if elapsed > time.Second {
		t.Errorf("last attempt should not backoff: took %v", elapsed)
	}
}
