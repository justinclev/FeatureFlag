package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	allowed := "http://allowed.com"
	h := CORS(allowed, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", allowed)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Header().Get("Access-Control-Allow-Origin") != allowed {
		t.Error("CORS header not set correctly")
	}
}
