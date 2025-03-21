package server

import (
	"context"
	"net/http"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/appenv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter/redis_ratelimiter"
	"github.com/rs/cors"
)

func (s *Server) RegisterRoutes(ctx context.Context) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", apiRouter(ctx, s)))

	rateLimitGlobal := middleware.RateLimiter(
		func(r *http.Request) (string, error) {
			// since we are using the RealIp() middleware
			// it should be safe to use r.RemoteAddr as limit key
			return r.RemoteAddr, nil
		},
		redis_ratelimiter.NewRedisSlidingWindowLimiter(
			ctx,
			s.rdb,
			ratelimiter.Config{
				PerTimeFrame: 60,
				TimeFrame:    time.Minute,
				KeyPrefix:    "global",
			},
		),
	)

	return middleware.MiddlewareChain(
		mux.ServeHTTP,
		corsMiddleware,
		s.LoggerInjector,
		// required for the rate limiter to function correctly and for logging
		middleware.RealIp(),
		middleware.RequestUUIDMiddleware,
		middleware.LocalizerInjector,
		middleware.RequestLogger,
		middleware.StripSlashes,
		rateLimitGlobal,
		middleware.Heartbeat,
	)
}

func corsMiddleware(next http.Handler) http.HandlerFunc {
	isStagOrLocal := appenv.IsStagOrLocal()
	if isStagOrLocal {
		o := cors.Options{
			Debug:            isStagOrLocal,
			AllowedOrigins:   []string{"http://127.0.0.1:8080"},
			AllowedMethods:   []string{"OPTIONS", "HEAD", "GET", "POST", "DELETE", "PUT", "PATCH"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		}
		return middleware.Cors(o)(next)
	}
	return next.ServeHTTP
}

func apiRouter(ctx context.Context, s *Server) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/v1/", http.StripPrefix("/v1", v1Router(ctx, s)))

	return middleware.MiddlewareChain(
		mux.ServeHTTP,
	)
}

func v1Router(ctx context.Context, s *Server) http.Handler {
	mux := http.NewServeMux()

	authRepo := s.NewAuthRepository()

	mux.Handle("/auth/", http.StripPrefix("/auth", authRouter(ctx, s, authRepo)))

	mux.Handle("/todo", todoRouter(s))
	mux.Handle("/todo/", todoRouter(s))

	if appenv.IsStagOrLocal() {
		mux.Handle("/dev-tools/", http.StripPrefix("/dev-tools", middleware.NoCache(devToolsRouter(s))))
	}

	return middleware.MiddlewareChain(
		mux.ServeHTTP,
	)
}
