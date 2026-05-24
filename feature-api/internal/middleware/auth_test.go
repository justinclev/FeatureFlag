package middleware_test

import (
    "bytes"
	"net/http"
	"net/http/httptest"
    "log/slog"
	"testing"

	"github.com/featureflags/feature-api/internal/middleware"
)

const testKey = "test-api-key"

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func applyAuth(next http.HandlerFunc) http.Handler {
	return middleware.APIKeyAuth(testKey, nil, http.HandlerFunc(next))
}

func TestAPIKeyAuth_ValidKey(t *testing.T) {
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/flags", nil)
	r.Header.Set("X-API-KEY", testKey)

	applyAuth(okHandler).ServeHTTP(rr, r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/flags", nil)

	applyAuth(okHandler).ServeHTTP(rr, r)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/flags", nil)
	r.Header.Set("X-API-KEY", "wrong-key")

	applyAuth(okHandler).ServeHTTP(rr, r)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAPIKeyAuth_WithLogger(t *testing.T) {
    var buf bytes.Buffer
    logger := slog.New(slog.NewTextHandler(&buf, nil))
    
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/flags", nil)
    
    h := middleware.APIKeyAuth(testKey, logger, http.HandlerFunc(okHandler))
	h.ServeHTTP(rr, r)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
    if !bytes.Contains(buf.Bytes(), []byte("unauthorized request")) {
        t.Error("expected log message")
    }
}
