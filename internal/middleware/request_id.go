package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type contextKey string

const requestIDKey contextKey = "requestID"

// RequestID - attaches an X-Request-ID header to the request context and response.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		ctx := context.WithValue(
			r.Context(),
			requestIDKey,
			requestID,
		)
		w.Header().Set(
			"X-Request-ID",
			requestID,
		)
		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)
	})
}

// generateRequestID - returns a random hex request id, or "unknown" if entropy fails.
func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}
