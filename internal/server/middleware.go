package server

import (
	"fmt"
	"net/http"
	"strings"

	AppEnv "github.com/Nidal-Bakir/go-todo-backend/internal/app_env"
	"github.com/Nidal-Bakir/go-todo-backend/internal/user"
)

func (s *Server) Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.Replace(r.Header.Get("Authorization"), "Bearer", "", 1))
		// TODO: also check if its a valid token (length, schema etc..)
		if token == "" {
			WriteError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
			return
		}

		userActions := user.NewActions(s.db.Queries)
		userModel, err := userActions.GetUserBySessionToken(r.Context(), token)
		if err != nil {
			if AppEnv.IsStagOrLocal() {
				WriteError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
			}
			WriteError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
			return
		}

		r = r.WithContext(user.ContextWithUser(r.Context(), userModel))
		h.ServeHTTP(w, r)
	})
}
