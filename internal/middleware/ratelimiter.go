package middleware

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
)

func RateLimiter(config ratelimiter.Config, limitKeyFn func(r *http.Request) string) func(next http.Handler) http.HandlerFunc {
	limiter := ratelimiter.NewTokenBucketLimiter(config)

	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			allow, backoffDuration := limiter.Allow(limitKeyFn(r))

			if !allow {
				w.Header().Add("Retry-After", strconv.Itoa(int(backoffDuration.Seconds())))
				// Request limit per ${config.TimeFrame}
				w.Header().Add("X-RateLimit-Limit ", fmt.Sprint(config.RequestsPerTimeFrame))
				utils.WriteError(r.Context(), w, http.StatusTooManyRequests, errors.New("Too Many Requests!"))
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

func LimitKeyFnUtil(r *http.Request, trustedHeaders ...string) string {
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
