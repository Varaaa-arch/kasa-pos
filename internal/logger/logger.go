package logger

import (
	"context"
	"log/slog"
	"os"
)

// contextKey is unexported to avoid collision with other packages.
type contextKey string

const requestIDKey contextKey = "request_id"

// Init initialises the global slog logger with a text handler.
// Call this once at application startup (main.go) before any logging.
//
// Output format:
//
//	time=2026-08-22T21:00:00Z level=INFO msg=checkout_completed event=checkout_completed request_id=req_xxx
func Init() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(handler))
}

// ContextWithRequestID stores a request ID in the context.
// Called by the RequestID middleware.
func ContextWithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFromContext retrieves the request ID stored by ContextWithRequestID.
// Returns empty string if not present.
func RequestIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(requestIDKey).(string)
	return v
}
