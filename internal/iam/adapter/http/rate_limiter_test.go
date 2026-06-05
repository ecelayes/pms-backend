package http

import (
	"net/http"
	"sync"
	"time"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestIPRateLimiter_AllowsUpToBurst(t *testing.T) {
	e := echo.New()
	limiter := NewIPRateLimiter(0.1, 3, time.Minute)
	handler := limiter.Middleware()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if err := handler(c); err != nil {
			t.Fatalf("expected ok on call %d, got %v", i, err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("call %d: expected 200, got %d", i+1, rec.Code)
		}
	}
}

func TestIPRateLimiter_Blocks429AfterBurst(t *testing.T) {
	e := echo.New()

	limiter := NewIPRateLimiter(0.1, 2, time.Minute)
	handler := limiter.Middleware()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	call := func() int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		_ = handler(c)
		return rec.Code
	}

	for i := 0; i < 2; i++ {
		if code := call(); code != http.StatusOK {
			t.Errorf("call %d: expected 200, got %d", i+1, code)
		}
	}
	if code := call(); code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", code)
	}
}

func TestIPRateLimiter_PerIPIsolation(t *testing.T) {
	e := echo.New()

	limiter := NewIPRateLimiter(0.1, 1, time.Minute)
	handler := limiter.Middleware()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	// First IP
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec1)
	c1.Request().RemoteAddr = "1.1.1.1:1234"
	_ = handler(c1)
	_ = handler(c1) // 2nd should be blocked for this IP

	// Second IP - should still work
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec2)
	c2.Request().RemoteAddr = "2.2.2.2:1234"
	if err := handler(c2); err != nil {
		t.Errorf("second IP should not be rate limited, got %v", err)
	}
}

func TestIPRateLimiter_RaceConditionConcurrent(t *testing.T) {
	limiter := NewIPRateLimiter(100, 100, time.Minute)
	handler := limiter.Middleware()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	const goroutines = 50
	const requestsPerGoroutine = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(gid int) {
			defer wg.Done()
			for r := 0; r < requestsPerGoroutine; r++ {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.RemoteAddr = "10.0.0.1:1234"
				rec := httptest.NewRecorder()
				c := echo.New().NewContext(req, rec)
				_ = handler(c)
			}
		}(g)
	}
	wg.Wait()
}

// TestIPRateLimiter_ConcurrentDifferentIPs isolates per-IP buckets under concurrent load.
func TestIPRateLimiter_ConcurrentDifferentIPs(t *testing.T) {
	limiter := NewIPRateLimiter(1, 1, time.Minute)
	handler := limiter.Middleware()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	const ips = 10
	const requestsPerIP = 5
	var wg sync.WaitGroup
	wg.Add(ips)

	for i := 0; i < ips; i++ {
		go func(ipNum int) {
			defer wg.Done()
			ip := "192.168.0." + string(rune('0'+ipNum%10))
			for r := 0; r < requestsPerIP; r++ {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.RemoteAddr = ip + ":1234"
				rec := httptest.NewRecorder()
				c := echo.New().NewContext(req, rec)
				_ = handler(c)
			}
		}(i)
	}
	wg.Wait()
}

func TestIPRateLimiter_Middleware_FallbackRemoteAddr(t *testing.T) {
	e := echo.New()
	limiter := NewIPRateLimiter(100, 200, time.Hour)
	e.Use(limiter.Middleware())
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	// Use a request without X-Forwarded-For and without a real IP header
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Del(echo.HeaderXRealIP)
	req.Header.Del(echo.HeaderXForwardedFor)
	req.RemoteAddr = "192.168.1.1:1234"
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestIPRateLimiter_GCRemovesStaleBuckets(t *testing.T) {
	limiter := &IPRateLimiter{
		buckets: make(map[string]*ipLimiter),
		r:       10,
		b:       20,
		ttl:     10 * time.Millisecond,
	}

	limiter.get("stale-ip")
	if len(limiter.buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(limiter.buckets))
	}

	time.Sleep(20 * time.Millisecond)
	limiter.gcOnce()

	if len(limiter.buckets) != 0 {
		t.Errorf("expected stale bucket to be GC'd, got %d buckets", len(limiter.buckets))
	}
}

func TestIPRateLimiter_GCKeepsFreshBuckets(t *testing.T) {
	limiter := &IPRateLimiter{
		buckets: make(map[string]*ipLimiter),
		r:       10,
		b:       20,
		ttl:     time.Hour,
	}

	limiter.get("fresh-ip")
	limiter.gcOnce()

	if len(limiter.buckets) != 1 {
		t.Errorf("expected fresh bucket to remain, got %d", len(limiter.buckets))
	}
}

func TestIPRateLimiter_GCLoopRunsAndStops(t *testing.T) {
	limiter := &IPRateLimiter{
		buckets: make(map[string]*ipLimiter),
		r:       10,
		b:       20,
		ttl:     5 * time.Millisecond,
		stop:    make(chan struct{}),
	}

	limiter.get("stale-ip")
	done := make(chan struct{})
	go func() {
		limiter.gcLoop(limiter.stop)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	close(limiter.stop)
	<-done

	limiter.mu.Lock()
	count := len(limiter.buckets)
	limiter.mu.Unlock()

	if count != 0 {
		t.Errorf("expected stale bucket to be GCd by loop, got %d buckets", count)
	}
}

func TestIPRateLimiter_Middleware_EmptyIP(t *testing.T) {
	e := echo.New()
	limiter := NewIPRateLimiter(100, 200, time.Hour)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Del(echo.HeaderXRealIP)
	req.Header.Del(echo.HeaderXForwardedFor)
	req.RemoteAddr = ""
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := limiter.Middleware()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	_ = handler(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200 with empty IP, got %d", rec.Code)
	}
}

func TestStripPort(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"192.168.1.1", "192.168.1.1"},
		{"192.168.1.1:8080", "192.168.1.1"},
		{"[::1]:8080", "[::1]"},
		{"", ""},
	}
	for _, tc := range tests {
		if got := stripPort(tc.input); got != tc.want {
			t.Errorf("stripPort(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
