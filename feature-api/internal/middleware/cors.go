package middleware

import (
	"net/http"
	"strings"
)

// CORS injects cross-origin headers.
func CORS(allowedOrigin string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// If origin matches exactly, or is the 127.0.0.1 equivalent of localhost
			isAllowed := origin == allowedOrigin
			if !isAllowed && strings.Contains(allowedOrigin, "localhost") {
				altOrigin := strings.Replace(allowedOrigin, "localhost", "127.0.0.1", 1)
				isAllowed = origin == altOrigin
			}

			if isAllowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-KEY")
			w.Header().Set("Vary", "Origin")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
