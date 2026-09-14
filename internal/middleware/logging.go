package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logging - logs method, path, and duration for each HTTP request.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf(
			"method=%s path=%s duration=%s",
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
	})
}
