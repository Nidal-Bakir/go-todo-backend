package server

import (
	"context"
	"net/http"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/appenv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth"
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
		middleware.RequestLoggerWithHeaderSkipFn(func(headerName string) bool {
			if headerName == "A-Installation" && appenv.IsStagOrLocal() {
				return false
			}
			return true
		}),
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

	registerAuthHandler(ctx, mux, s, authRepo)
	registerInstallationHandler(ctx, mux, authRepo)

	registerTodoHandler(ctx, mux, s)

	if appenv.IsStagOrLocal() {
		mux.Handle("/dev-tools/", http.StripPrefix("/dev-tools", devToolsRouter(s)))
	}

	return middleware.MiddlewareChain(
		mux.ServeHTTP,
	)
}

// handel: /auth/
func registerAuthHandler(ctx context.Context, mux *http.ServeMux, s *Server, authRepo auth.Repository) {
	mux.Handle("/auth/", http.StripPrefix("/auth", authRouter(ctx, s, authRepo)))
}

// handel: /installation/
func registerInstallationHandler(ctx context.Context, mux *http.ServeMux, authRepo auth.Repository) {
	mux.Handle("/installation/", http.StripPrefix("/installation", installationRouter(ctx, authRepo)))
}

// handel: /todo and /todo/
func registerTodoHandler(ctx context.Context, mux *http.ServeMux, s *Server) {
	h := todoRouter(ctx, s)
	mux.Handle("/todo", h)
	mux.Handle("/todo/", h)
}
