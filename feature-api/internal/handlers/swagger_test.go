package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterSwagger(t *testing.T) {
	mux := http.NewServeMux()
	RegisterSwagger(mux)

	req := httptest.NewRequest("GET", "/openapi.yaml", nil)
	rw := httptest.NewRecorder()
	mux.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK || rw.Header().Get("Content-Type") != "application/yaml" {
		t.Error("/openapi.yaml not served correctly")
	}

	req2 := httptest.NewRequest("GET", "/docs", nil)
	rw2 := httptest.NewRecorder()
	mux.ServeHTTP(rw2, req2)
	if rw2.Code != http.StatusOK || rw2.Header().Get("Content-Type") != "text/html" {
		t.Error("/docs not served correctly")
	}
}
