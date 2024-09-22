package middleware

import (
	"net/http"
	"time"
)

// Sunset set Deprecation/Sunset header to response
// This can be used to enable Sunset in a route or a route group
// NOTE: the sunsetAt Time should be in UTC time zone
// For more: https://www.rfc-editor.org/rfc/rfc8594.html
func Sunset(sunsetAt time.Time, links ...string) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !sunsetAt.IsZero() {
				w.Header().Set("Sunset", sunsetAt.Format(http.TimeFormat))
				w.Header().Set("Deprecation", sunsetAt.Format(http.TimeFormat))

				for _, link := range links {
					w.Header().Add("Link", link)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
