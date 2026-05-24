package middleware

import (
	"crypto/subtle"
	"net/http"
)

// APIKeyAuth returns a middleware that validates the "X-API-KEY" header.
func APIKeyAuth(expectedKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-KEY")
		
		// Principal Security: Limit API Key entropy to prevent timing/OOM attacks on the header.
		if len(key) == 0 || len(key) > 128 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if subtle.ConstantTimeCompare([]byte(key), []byte(expectedKey)) != 1 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
