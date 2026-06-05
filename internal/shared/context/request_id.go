package context

import (
	"context"
	"errors"

	"github.com/labstack/echo/v4"
)

type requestIDKey struct{}

// ErrNoRequestID is returned when no request_id is found in the context.
var ErrNoRequestID = errors.New("no request_id in context")

const requestIDHeader = "X-Request-ID"

// WithRequestID stores the request_id in the context.
// The request_id propagates from HTTP requests to async goroutines.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// RequestIDFromContext extracts the request_id from the context.
// Returns ErrNoRequestID if no request_id is present.
func RequestIDFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", ErrNoRequestID
	}
	if v, ok := ctx.Value(requestIDKey{}).(string); ok && v != "" {
		return v, nil
	}
	return "", ErrNoRequestID
}

// RequestIDFromEcho extracts the request_id from an Echo context, falling back
// to the X-Request-ID response header. Returns an empty string if not present.
func RequestIDFromEcho(c echo.Context) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Get("request_id").(string); ok && v != "" {
		return v
	}
	return c.Response().Header().Get(requestIDHeader)
}
