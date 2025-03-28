package middleware

import (
	"context"
	"net/http"
)

// WithValue is a middleware that sets a given key/value in a context chain.
func WithValue(key, val any) func(next http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), key, val))
			next.ServeHTTP(w, r)
		}
	}
}
