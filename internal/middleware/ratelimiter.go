package middleware

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/Nidal-Bakir/go-todo-backend/internal/l10n"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
)

func RateLimiter(limitKeyFn func(r *http.Request) string, limiter ratelimiter.Limiter) func(next http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			allow, backoffDuration := limiter.Allow(ctx, limitKeyFn(r))

			if !allow {
				local, ok := l10n.LocalizerFromContext(ctx)
				utils.AssertDev(ok, "LocalizerFromContext could not find the localizer in the context tree!")

				w.Header().Add("Retry-After", strconv.Itoa(int(math.Ceil(backoffDuration.Abs().Seconds()))))
				// Request limit per ${config.TimeFrame}
				w.Header().Add("X-RateLimit-Limit", fmt.Sprint(limiter.Config().PerTimeFrame))
				utils.WriteError(r.Context(), w, http.StatusTooManyRequests, errors.New(local.GetWithId("too_many_requests")))
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}
