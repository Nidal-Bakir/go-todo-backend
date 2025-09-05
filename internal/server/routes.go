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

	authRepo := s.NewAuthRepository()

	mux.Handle("/api/", http.StripPrefix("/api", apiRouter(ctx, s, authRepo)))
	mux.Handle("/", webRouter(ctx, authRepo))

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
		s.LoggerInjector,
		corsMiddleware,
		// required for the rate limiter to function correctly and for logging
		middleware.RealIp(),
		middleware.RequestUUIDMiddleware,
		middleware.LocalizerInjector,
		middleware.If(
			appenv.IsLocal(),
			middleware.RequestLoggerWithHeaderSkipFn(
				func(headerName string) bool {
					if (headerName == "A-Installation" ||
						headerName == "Authorization" ||
						headerName == "Postman-Token") &&
						appenv.IsStagOrLocal() {
						return false
					}
					return true
				},
			),
		),
		middleware.StripSlashes,
		rateLimitGlobal,
		middleware.Heartbeat,
		middleware.CSRFProtection(FrontendDomains...),
	)
}

func corsMiddleware(next http.Handler) http.HandlerFunc {
	o := cors.Options{
		Debug:            appenv.IsLocal(),
		AllowedOrigins:   FrontendDomains,
		AllowedMethods:   []string{"OPTIONS", "HEAD", "GET", "POST", "DELETE", "PUT", "PATCH"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Accept", "Accept-Language"},
		AllowCredentials: true,
		MaxAge:           10, // 10 sec
	}
	return middleware.Cors(o)(next)
}

func apiRouter(ctx context.Context, s *Server, authRepo auth.Repository) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/v1/", http.StripPrefix("/v1", v1Router(ctx, s, authRepo)))

	return middleware.MiddlewareChain(
		mux.ServeHTTP,
	)
}

func v1Router(ctx context.Context, s *Server, authRepo auth.Repository) http.Handler {
	mux := http.NewServeMux()

	registerAuthHandler(ctx, mux, s, authRepo)
	registerInstallationHandler(ctx, mux, authRepo)

	registerTodoHandler(ctx, mux, s, authRepo)

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
//
// Needs: Auth
func registerTodoHandler(ctx context.Context, mux *http.ServeMux, s *Server, authRepo auth.Repository) {
	h := middleware.MiddlewareChain(
		todoRouter(ctx, s).ServeHTTP,
		Auth(authRepo),
	)

	mux.Handle("/todo", h)
	mux.Handle("/todo/", h)
}
