package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recovery returns a middleware that recovers from panics and logs the error.
func Recovery(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"method", r.Method,
					"path", r.URL.Path,
				)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
