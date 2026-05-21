package middleware
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/featureflags/feature-api/internal/middleware"
)

const testKey = "test-api-key"

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func applyAuth(next http.HandlerFunc) http.Handler {
	return middleware.APIKeyAuth(testKey, http.HandlerFunc(next))
}

func TestAPIKeyAuth_ValidKey(t *testing.T) {
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/flags", nil)
	r.Header.Set("X-API-Key", testKey)

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
	r.Header.Set("X-API-Key", "wrong-key")

	applyAuth(okHandler).ServeHTTP(rr, r)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestAPIKeyAuth_HealthSkipsAuth(t *testing.T) {
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	// no X-API-Key header

	applyAuth(okHandler).ServeHTTP(rr, r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected /health to bypass auth, got %d", rr.Code)
	}
}
