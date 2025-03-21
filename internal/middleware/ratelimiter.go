package middleware

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/apiutils"
	"github.com/rs/zerolog"
)

func RateLimiter(limitKeyFn func(r *http.Request) (string, error), limiter ratelimiter.Limiter) func(next http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			key, err := limitKeyFn(r)
			if err != nil {
				zlog := zerolog.Ctx(ctx)
				apiutils.WriteError(ctx, w, http.StatusInternalServerError, err)
				zlog.Err(err).Msg("error while getting the limit key for the rate limiter")
				return
			}

			allow, backoffDuration := limiter.Allow(ctx, key)

			if !allow {
				w.Header().Add("Retry-After", strconv.Itoa(int(math.Ceil(backoffDuration.Abs().Seconds()))))
				// Request limit per ${config.TimeFrame}
				w.Header().Add("X-RateLimit-Limit", fmt.Sprint(limiter.Config().PerTimeFrame))
				apiutils.WriteError(ctx, w, http.StatusTooManyRequests, apperr.ErrTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}
