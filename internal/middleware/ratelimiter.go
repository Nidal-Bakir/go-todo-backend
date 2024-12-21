package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
)

func RateLimiter(limitKeyFn func(r *http.Request) string, limiter ratelimiter.Limiter) func(next http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			allow, backoffDuration := limiter.Allow(limitKeyFn(r))

			if !allow {
				w.Header().Add("Retry-After", strconv.Itoa(int(backoffDuration.Seconds())))
				// Request limit per ${config.TimeFrame}
				w.Header().Add("X-RateLimit-Limit ", fmt.Sprint(limiter.Config().RequestsPerTimeFrame))
				utils.WriteError(r.Context(), w, http.StatusTooManyRequests, errors.New("Too Many Requests!"))
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}
