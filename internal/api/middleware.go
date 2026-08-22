package api

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	applogger "pos-system/internal/logger"
)

// RequestID injects a unique request ID into each request context and sets the
// X-Request-ID response header. The ID is also available to handlers via
// RequestIDFromContext so it can be embedded in error responses.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}

		ctx := applogger.ContextWithRequestID(r.Context(), id)
		w.Header().Set("X-Request-ID", id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// newRequestID generates a short, URL-safe request ID in the form "req_<8hex>".
func newRequestID() string {
	return fmt.Sprintf("req_%08x", rand.New(rand.NewSource(time.Now().UnixNano())).Uint32())
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Idempotency-Key")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		slog.InfoContext(
			r.Context(),
			"http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start).String(),
			"request_id", applogger.RequestIDFromContext(r.Context()),
		)
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(
					w,
					"internal server error",
					http.StatusInternalServerError,
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func Chain(
	handler http.Handler,
	middlewares ...func(http.Handler) http.Handler,
) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}
