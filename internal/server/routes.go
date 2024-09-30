package server

import (
	"net/http"

	"github.com/Nidal-Bakir/go-todo-backend/internal/AppEnv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", apiRouter(s)))

	return middleware.MiddlewareChain(
		mux.ServeHTTP,
		s.LoggerInjector,
		middleware.RequestUUIDMiddleware,
		middleware.LocalizerInjector,
		middleware.RequestLogger,
		middleware.StripSlashes,
		middleware.AllowContentType("application/json"),
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

	mux.Handle("/todo", middleware.Throttle(1)(todoRouter(s)))
	mux.Handle("/todo/", todoRouter(s))

	if AppEnv.IsStagOrLocal() {
		mux.Handle("/dev-tools", http.StripPrefix("/dev-tools", middleware.NoCache(devToolsRouter(s))))
	}

	return middleware.MiddlewareChain(
		mux.ServeHTTP,
	)
}
