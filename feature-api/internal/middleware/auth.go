package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
)

// APIKeyAuth protects all routes except /health by requiring a valid X-API-Key header.
// Returns 401 when the header is absent, 403 when the key is wrong.
func APIKeyAuth(apiKey string, next http.Handler) http.Handler {
	keyBytes := []byte(apiKey)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		got := r.Header.Get("X-API-Key")
		if got == "" {
			writeAuthError(w, http.StatusUnauthorized, "missing API key")
			return
		}
		if subtle.ConstantTimeCompare([]byte(got), keyBytes) != 1 {
			writeAuthError(w, http.StatusForbidden, "invalid API key")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeAuthError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
