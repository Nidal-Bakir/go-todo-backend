package server

import (
	"context"
	"net/http"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/AppEnv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter/redis_ratelimiter"
)

func (s *Server) RegisterRoutes(ctx context.Context) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", apiRouter(s)))

	rateLimitGlobal := middleware.RateLimiter(
		func(r *http.Request) string {
			// since we are using the RealIp() middleware
			// it should be safe to use r.RemoteAddr as limit key
			return r.RemoteAddr
		},
		redis_ratelimiter.NewRedisTokenBucketLimiter(
			ctx,
			s.rdb,
			ratelimiter.Config{
				Enabled:      true,
				PerTimeFrame: 5,
				TimeFrame:    time.Minute,
				KeyPrefix:    "global",
			},
		),
	)

	return middleware.MiddlewareChain(
		mux.ServeHTTP,
		s.LoggerInjector,
		// required for the rate limiter to function correctly and for logging
		middleware.RealIp(),
		middleware.RequestUUIDMiddleware,
		middleware.LocalizerInjector,
		middleware.RequestLogger,
		middleware.StripSlashes,
		middleware.AllowContentType("application/json"),
		rateLimitGlobal,
		middleware.Heartbeat,
	)
}

func apiRouter(s *Server) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/v1/", http.StripPrefix("/v1", v1Router(s)))

	return middleware.MiddlewareChain(
		mux.ServeHTTP,
	)
}

func v1Router(s *Server) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/auth/", http.StripPrefix("/auth", authRouter(s)))

	mux.Handle("/todo", todoRouter(s))
	mux.Handle("/todo/", todoRouter(s))

	if AppEnv.IsStagOrLocal() {
		mux.Handle("/dev-tools/", http.StripPrefix("/dev-tools", middleware.NoCache(devToolsRouter(s))))
	}

	return middleware.MiddlewareChain(
		mux.ServeHTTP,
	)
}
