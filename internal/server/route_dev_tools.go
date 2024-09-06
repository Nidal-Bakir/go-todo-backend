package server

import (
	"net/http"
)

func devToolsRouter(s *Server) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.healthHandler)
	return mux
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	WriteJson(r.Context(), w, http.StatusOK, s.db.Health(r.Context()))
}
