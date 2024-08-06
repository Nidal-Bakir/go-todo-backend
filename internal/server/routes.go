package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (s *Server) RegisterRoutes() http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", s.HelloWorldHandler)
	mux.HandleFunc("/health", s.healthHandler)

	return mux
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {

	WriteError(w,
		500,
		fmt.Errorf("some error"),
		fmt.Errorf("error1"), fmt.Errorf("error2"),
	)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health(r.Context()))

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	w.Write(jsonResp)
}
