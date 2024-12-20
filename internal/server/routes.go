package server

import (
	"net/http"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/AppEnv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", apiRouter(s)))

	rateLimitGlobal := middleware.RateLimiter(
		func(r *http.Request) string {
			// since we are using the RealIp() middleware
			// it should be safe to use r.RemoteAddr as limit key
			return r.RemoteAddr
		},
		ratelimiter.NewTokenBucketLimiter(
			ratelimiter.Config{
				Enabled:              true,
				RequestsPerTimeFrame: 5,
				TimeFrame:            time.Minute,
			},
		),
	)

	return middleware.MiddlewareChain(
		mux.ServeHTTP,
		s.LoggerInjector,
		middleware.RealIp(), // required for the rate limiter to function correctly
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
