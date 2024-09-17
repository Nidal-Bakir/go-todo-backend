package middleware

import (
	"net/http"
)

// RequestSize is a middleware that will limit request sizes to a specified
// number of bytes. It uses MaxBytesReader to do so.
func RequestSize(bytes int64) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, bytes)
			next.ServeHTTP(w, r)
		})
	}
}
