package server

import (
	"context"
	"net/http"
)

func todoRouter(_ context.Context, s *Server) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /todo", createTodo())
	mux.HandleFunc("GET /todo", todoIndex())
	mux.HandleFunc("GET /todo/{id}", todoShow())
	return mux
}

func createTodo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, http.StatusBadRequest, err)
			return
		}
	}
}

func todoIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, http.StatusBadRequest, err)
			return
		}
	}
}
func todoShow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, http.StatusBadRequest, err)
			return
		}
	}
}
