package server

import (
	"expvar"
	"net/http"
	"net/http/pprof"
)

func devToolsRouter(s *Server) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.healthHandler)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/pprof/", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/pprof", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/", http.StatusMovedPermanently)
	})

	mux.HandleFunc("/pprof/*", pprof.Index)
	mux.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/pprof/profile", pprof.Profile)
	mux.HandleFunc("/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/pprof/trace", pprof.Trace)
	mux.Handle("/vars", expvar.Handler())

	mux.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))
	mux.Handle("/pprof/mutex", pprof.Handler("mutex"))
	mux.Handle("/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/pprof/block", pprof.Handler("block"))
	mux.Handle("/pprof/allocs", pprof.Handler("allocs"))

	return mux
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	WriteJson(r.Context(), w, http.StatusOK, s.db.Health(r.Context()))
}
