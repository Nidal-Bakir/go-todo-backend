package middleware

import (
	"net"
	"net/http"

	"github.com/rs/zerolog"
)

func RealIp(trustedHeaders ...string) func(next http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := *zerolog.Ctx(r.Context())
			
			r.RemoteAddr = realIp(r, trustedHeaders...)
			
			ctx = log.With().Str("RealIP", r.RemoteAddr).Logger().WithContext(ctx)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}

}

func realIp(r *http.Request, trustedHeaders ...string) string {
	for _, trustedHeader := range trustedHeaders {
		if headerVal := r.Header.Get(trustedHeader); headerVal != "" {
			if net.ParseIP(headerVal) != nil {
				return canonicalizeIP(headerVal)
			}
		}
	}

	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return canonicalizeIP(ip)
	}

	return r.RemoteAddr
}
