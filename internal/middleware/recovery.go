package middleware

import (
	"log"
	"net/http"
)

// Recovery - catches panics and returns a JSON 500 instead of crashing the process.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf(
					"panic recovered: %v",
					err,
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(
					`{"code":500,"type":"InternalError","message":"internal server error"}`,
				))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
