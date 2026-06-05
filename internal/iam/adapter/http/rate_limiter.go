package http

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// ipLimiter tracks per-IP token buckets.
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter returns an Echo middleware that limits requests per source IP.
//
// SECURITY: Rate limiting on auth endpoints mitigates:
//   - Brute-force login attempts
//   - Credential stuffing
//   - Password-reset flooding
//   - Token enumeration
//
// We use golang.org/x/time/rate (token bucket algorithm).
// Each IP gets its own bucket. Stale buckets are cleaned up periodically
// to prevent memory growth.
type IPRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*ipLimiter
	r       rate.Limit
	b       int
	ttl     time.Duration
	stop    chan struct{}
}

func NewIPRateLimiter(perSecond float64, burst int, ttl time.Duration) *IPRateLimiter {
	l := &IPRateLimiter{
		buckets: make(map[string]*ipLimiter),
		r:       rate.Limit(perSecond),
		b:       burst,
		ttl:     ttl,
	}
	stopCh := make(chan struct{})
	l.stop = stopCh
	go l.gcLoop(stopCh)
	return l
}

func (l *IPRateLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	bucket, ok := l.buckets[ip]
	if !ok {
		bucket = &ipLimiter{limiter: rate.NewLimiter(l.r, l.b)}
		l.buckets[ip] = bucket
	}
	bucket.lastSeen = time.Now()
	return bucket.limiter
}

func (l *IPRateLimiter) gcOnce() {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	for ip, b := range l.buckets {
		if now.Sub(b.lastSeen) > l.ttl {
			delete(l.buckets, ip)
		}
	}
}

func (l *IPRateLimiter) gcLoop(stop <-chan struct{}) {
	t := time.NewTicker(l.ttl)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			l.gcOnce()
		}
	}
}

// stripPort removes a trailing :port from an IP literal if present.
// RealIP usually strips the port, but we keep this as a defensive fallback.
func stripPort(ip string) string {
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		return ip[:idx]
	}
	return ip
}

func (l *IPRateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			if ip == "" {
				ip = c.Request().RemoteAddr
			}
			ip = stripPort(ip)
			if !l.get(ip).Allow() {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "too many requests",
				})
			}
			return next(c)
		}
	}
}
