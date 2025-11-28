package middleware

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/resutils"
	"github.com/rs/zerolog"
)

func RateLimiterWithOptionalLimit(limitKeyFn func(r *http.Request) (key string, shouldRateLimit bool, err error), limiter ratelimiter.Limiter) func(next http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			key, shouldRateLimit, err := limitKeyFn(r)
			if err != nil {
				zlog := zerolog.Ctx(ctx)
				resutils.WriteError(ctx, w, r, http.StatusInternalServerError, err)
				zlog.Err(err).Msg("error while getting the limit key for the rate limiter")
				return
			}

			if shouldRateLimit {
				allow, backoffDuration := limiter.Allow(ctx, key)
				if !allow {
					w.Header().Add("Retry-After", strconv.Itoa(int(math.Ceil(backoffDuration.Abs().Seconds()))))
					// Request limit per ${config.TimeFrame}
					w.Header().Add("X-RateLimit-Limit", fmt.Sprint(limiter.Config().PerTimeFrame))
					resutils.WriteError(ctx, w, r, http.StatusTooManyRequests, apperr.ErrTooManyRequests)
					return
				}
			}

			next.ServeHTTP(w, r)
		}
	}
}

func RateLimiter(limitKeyFn func(r *http.Request) (string, error), limiter ratelimiter.Limiter) func(next http.Handler) http.HandlerFunc {
	return RateLimiterWithOptionalLimit(
		func(r *http.Request) (string, bool, error) {
			key, err := limitKeyFn(r)
			return key, true, err
		},
		limiter,
	)
}
