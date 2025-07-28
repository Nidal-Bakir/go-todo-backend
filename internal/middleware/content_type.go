package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/mimes"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/resutils"
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

			resutils.WriteError(r.Context(), w, r, http.StatusUnsupportedMediaType, errors.New("did you set the Content-Type header correctly?"))
		})
	}
}

// Only allow the content type application/x-www-form-urlencoded
//
// Use AllowContentType, to construct custom allowed list of content types
func ACT_app_x_www_form_urlencoded(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		AllowContentType(mimes.App_x_www_form_urlencoded)(next).ServeHTTP(w, r)
	}
}

// Only allow the content type multipart/form-data
//
// Use AllowContentType, to construct custom allowed list of content types
func ACT_multipart_form_data(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		AllowContentType(mimes.Multipart_form_data)(next).ServeHTTP(w, r)
	}
}

// Only allow the content type application/json
//
// Use AllowContentType, to construct custom allowed list of content types
func ACT_app_json(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		AllowContentType(mimes.App_json)(next).ServeHTTP(w, r)
	}
}
