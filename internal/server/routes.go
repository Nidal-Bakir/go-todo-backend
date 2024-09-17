package server

import (
	"net/http"

	"github.com/Nidal-Bakir/go-todo-backend/internal/AppEnv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", v1Router(s)))

	return middleware.MiddlewareChain(
		mux.ServeHTTP,
		s.LoggerInjector,
		middleware.RequestUUIDMiddleware,
		middleware.LocalizerInjector,
		middleware.RequestLogger,
		middleware.StripSlashes,
		middleware.Heartbeat,
	)
}

func v1Router(s *Server) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/auth/", http.StripPrefix("/auth", authRouter(s)))

	mux.Handle("/todo", s.Auth(todoRouter(s)))
	mux.Handle("/todo/", s.Auth(todoRouter(s)))

	if AppEnv.IsStagOrLocal() {
		mux.Handle("/dev-tools", http.StripPrefix("/dev-tools", devToolsRouter(s)))
		mux.Handle("/dev-tools", http.StripPrefix("/dev-tools", devToolsRouter(s)))
	}

	return mux
}
