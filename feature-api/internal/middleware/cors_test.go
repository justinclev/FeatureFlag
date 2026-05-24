package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	allowed := "http://allowed.com"
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := CORS(allowed)(next)

	// GET with allowed origin
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", allowed)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Header().Get("Access-Control-Allow-Origin") != allowed {
		t.Error("CORS header not set correctly for allowed origin")
	}

	// GET with denied origin
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://denied.com")
	rw = httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("CORS header set for denied origin")
	}

	// OPTIONS (preflight)
	req = httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", allowed)
	rw = httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS, got %d", rw.Code)
	}
}
