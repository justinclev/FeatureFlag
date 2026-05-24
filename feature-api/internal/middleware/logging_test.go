package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogging(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	h := Logging(logger)(next)
	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusTeapot {
		t.Error("logging middleware did not pass status")
	}
}

func TestLogging_HotPath(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := Logging(logger)(next)
    
    // Path ending in /evaluate should skip logging logic (though it still calls next)
	req := httptest.NewRequest("POST", "/api/flags/foo/evaluate", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Error("hot path skip failed")
	}
}

func TestLogging_DisabledLevel(t *testing.T) {
    // Logger set to Error level, so Info logs (default) are disabled
    logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := Logging(logger)(next)
	req := httptest.NewRequest("GET", "/other", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Error("disabled level check failed")
	}
}
