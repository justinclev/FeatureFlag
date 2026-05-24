package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// responseRecorder wraps http.ResponseWriter to capture the written status code.
type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (rr *responseRecorder) WriteHeader(status int) {
	rr.status = status
	rr.ResponseWriter.WriteHeader(status)
}

// Logging returns a middleware that emits a structured slog line per request.
// Optimized: Skips logging for the evaluation hot path and checks level before processing.
func Logging(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			
			// Principal Optimization: Skips standard logging for the high-frequency evaluation path.
			if strings.HasSuffix(path, "/evaluate") {
				next.ServeHTTP(w, r)
				return
			}

			if !logger.Enabled(r.Context(), slog.LevelInfo) {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			
			logger.Info("request",
				"method", r.Method,
				"path", path,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
