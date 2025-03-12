package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

func Cors(options cors.Options) func(next http.Handler) http.HandlerFunc {
	c := cors.New(options)
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			c.Handler(next).ServeHTTP(w, r)
		}
	}
}
