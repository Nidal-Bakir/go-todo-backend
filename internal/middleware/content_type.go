package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
)

// SetHeader is a convenience handler to set a response header key/value
func SetHeader(key, value string) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(key, value)
			next.ServeHTTP(w, r)
		})
	}
}

// AllowContentType enforces a whitelist of request Content-Types otherwise responds
// with a 415 Unsupported Media Type status.
func AllowContentType(contentTypes ...string) func(http.Handler) http.HandlerFunc {
	allowedContentTypes := make(map[string]struct{}, len(contentTypes))
	for _, ctype := range contentTypes {
		allowedContentTypes[strings.TrimSpace(strings.ToLower(ctype))] = struct{}{}
	}

	return func(next http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength == 0 {
				// Skip check for empty content body
				next.ServeHTTP(w, r)
				return
			}

			s := strings.ToLower(strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0]))

			if _, ok := allowedContentTypes[s]; ok {
				next.ServeHTTP(w, r)
				return
			}

			utils.WriteError(r.Context(), w, http.StatusUnsupportedMediaType, errors.New("Did you set the Content-Type header correctly?"))
		})
	}
}