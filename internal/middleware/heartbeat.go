package middleware

import (
	"net/http"
)

func Heartbeat(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if (r.Method == "GET" || r.Method == "HEAD") && (r.URL.Path == "/ping") {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("pong"))
			return
		}
		next.ServeHTTP(w, r)
	}
}
