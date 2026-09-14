package middleware

import (
	"encoding/json"
	"net/http"
)

const validAPIKey = "apitest"

type APIResponse struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

// APIKey - rejects requests missing or using an invalid api_key header.
func APIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("api_key")
		if apiKey == "" {
			writeAuthError(w, http.StatusUnauthorized, "Unauthorized", "api_key is required")
			return
		}
		if apiKey != validAPIKey {
			writeAuthError(w, http.StatusForbidden, "Forbidden", "invalid api_key")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// writeAuthError - writes a JSON auth error payload with the given HTTP status.
func writeAuthError(w http.ResponseWriter, status int, errorType string, message string) {
	response := APIResponse{
		Code:    status,
		Type:    errorType,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}
