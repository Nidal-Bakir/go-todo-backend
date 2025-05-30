package middleware

import "net/http"

// Maybe middleware will allow you to change the flow of the middleware stack execution depending on return
// value of maybeFn(request). This is useful for example if you'd like to skip a middleware handler if
// a request does not satisfy the maybeFn logic.
func Maybe(mw func(http.Handler) http.HandlerFunc, maybeFn func(r *http.Request) bool) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if maybeFn(r) {
				mw(next).ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		}
	}
}

func If(con bool, mw func(http.Handler) http.HandlerFunc) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		if con {
			return mw(next).ServeHTTP
		}
		return next.ServeHTTP
	}
}
