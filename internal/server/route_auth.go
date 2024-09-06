package server

import "net/http"

func authRouter(s *Server) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /create-account", s.createAccount)
	return mux
}

func (s *Server) createAccount(w http.ResponseWriter, r *http.Request) {

}
