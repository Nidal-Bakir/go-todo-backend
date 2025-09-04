package middleware

import (
	"errors"
	"net"
	"net/http"
	"net/netip"

	"github.com/Nidal-Bakir/go-todo-backend/internal/tracker"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/resutils"
	"github.com/rs/zerolog"
)

func RealIp(trustedIpHeaders ...string) func(next http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			zlog := *zerolog.Ctx(ctx)

			r.RemoteAddr = RealIpFromRequest(r, trustedIpHeaders...)

			zlog = zlog.With().Str("client_ip", r.RemoteAddr).Logger()
			ctx = zlog.WithContext(ctx)

			requestIpAddres, err := netip.ParseAddr(r.RemoteAddr)
			if err != nil {
				zlog.Err(err).Msg("RealIp middleware: can not parse the remoteAddr using netip pkg")
				resutils.WriteError(ctx, w, r, http.StatusBadRequest, errors.New("can not parse the RemoteAddr"))
				return
			}
			ctx = tracker.ContextWithReqIP(ctx, requestIpAddres)

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
