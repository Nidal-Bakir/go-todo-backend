package server

import (
	"net/http"
	"strings"

	"github.com/Nidal-Bakir/go-todo-backend/internal/user"
)

func (s *Server) Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.Replace(r.Header.Get("Authorization"), "Bearer", "", 1))
		if token == "" {
			WriteJson(w, http.StatusUnauthorized, map[string]string{"": "Unauthorized"})
			return
		}

		userActions := user.NewActions(s.db.Queries)
		_, err := userActions.GetUserBySessionToken(r.Context(), token)
		if err != nil {

		}

		h.ServeHTTP(w, r)
	})
}
