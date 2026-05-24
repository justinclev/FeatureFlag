package middleware

import "net/http"

// Middleware defines a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain provides a way to nest multiple middlewares.
// The first middleware in the slice will be the outermost one.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// BodyLimit returns a middleware that limits the request body size.
func BodyLimit(maxSize int64) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			next.ServeHTTP(w, r)
		})
	}
}

// Wrap converts a standard middleware function with extra args into a Middleware type.
func Wrap(logger any, fn func(any, http.Handler) http.Handler) Middleware {
	return func(next http.Handler) http.Handler {
		return fn(logger, next)
	}
}
