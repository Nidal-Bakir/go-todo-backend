package middleware

import (
	"net"
	"net/http"

	"github.com/rs/zerolog"
)

func RealIp(trustedIpHeaders ...string) func(next http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := *zerolog.Ctx(r.Context())

			r.RemoteAddr = RealIpFromRequest(r, trustedIpHeaders...)

			ctx = log.With().Str("client_ip", r.RemoteAddr).Logger().WithContext(ctx)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

func RealIpFromRequest(r *http.Request, trustedIpHeaders ...string) string {
	for _, trustedIpHeader := range trustedIpHeaders {
		if headerVal := r.Header.Get(trustedIpHeader); headerVal != "" {
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
