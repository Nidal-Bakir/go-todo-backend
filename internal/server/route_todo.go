package server

import "net/http"

func todoRouter(s *Server) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /todo", s.createTodo)
	mux.HandleFunc("GET /todo", s.todoIndex)
	mux.HandleFunc("GET /todo/{id}", s.todoShow)
	return mux
}

func (s *Server) createTodo(w http.ResponseWriter, r *http.Request) {
	WriteJson(r.Context(), w, http.StatusOK, map[string]string{"createTodo": "createTodo"})
}
func (s *Server) todoIndex(w http.ResponseWriter, r *http.Request) {
	WriteJson(r.Context(), w, http.StatusOK, map[string]string{"todoIndex": "todoIndex"})
}

func (s *Server) todoShow(w http.ResponseWriter, r *http.Request) {
	WriteJson(r.Context(), w, http.StatusOK, map[string]string{"todoShow": "todoShow"})
}
