package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChain(t *testing.T) {
	order := ""
	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order += "1"
			next.ServeHTTP(w, r)
		})
	}
	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order += "2"
			next.ServeHTTP(w, r)
		})
	}

	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order += "H"
	}), m1, m2)

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	if order != "12H" {
		t.Errorf("expected chain order 12H, got %s", order)
	}
}

func TestBodyLimit(t *testing.T) {
	h := BodyLimit(5)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Under limit
	req := httptest.NewRequest("POST", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rw.Code)
	}
}

func TestWrap(t *testing.T) {
    wrapped := Wrap("foo", func(logger any, next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if logger.(string) == "foo" {
                w.WriteHeader(http.StatusAccepted)
            }
            next.ServeHTTP(w, r)
        })
    })
    
    h := wrapped(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
    rw := httptest.NewRecorder()
    h.ServeHTTP(rw, httptest.NewRequest("GET", "/", nil))
    if rw.Code != http.StatusAccepted {
        t.Errorf("Wrap did not pass logger correctly")
    }
}
