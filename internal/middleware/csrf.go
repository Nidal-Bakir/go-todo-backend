package middleware

import "net/http"

func CSRFProtection(trustedOrigin ...string) func(next http.Handler) http.HandlerFunc {
	csrf := http.NewCrossOriginProtection()
	for _, origin := range trustedOrigin {
		csrf.AddTrustedOrigin(origin)
	}
	return func(next http.Handler) http.HandlerFunc {
		return csrf.Handler(next).ServeHTTP
	}
}
