package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type ctxKey int

const requestIDKey ctxKey = 1

// RequestIDFrom exposes the request ID to handlers and error writers.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// WithRequestID assigns a server-generated ID to every request; client
// values are never trusted as identifiers, only echoed for correlation.
func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 8)
		if _, err := rand.Read(buf); err != nil {
			http.Error(w, "id source unavailable", http.StatusInternalServerError)
			return
		}
		id := hex.EncodeToString(buf)
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(
			context.WithValue(r.Context(), requestIDKey, id)))
	})
}

// WithRecovery converts a handler panic into the stable error envelope
// instead of tearing down the connection with no diagnosis.
func WithRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				WriteError(w, r, http.StatusInternalServerError,
					"internal", "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Chain applies middlewares outermost-first around the handler.
func Chain(h http.Handler, outer ...func(http.Handler) http.Handler) http.Handler {
	for i := len(outer) - 1; i >= 0; i-- {
		h = outer[i](h)
	}
	return h
}
