package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Nidal-Bakir/go-todo-backend/internal/AppEnv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/user"
	"github.com/jackc/pgx/v5"
)

func (s *Server) Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.Replace(r.Header.Get("Authorization"), "Bearer", "", 1))
		// TODO: also check if its a valid token (length, schema etc..)
		if token == "" {
			WriteError(r.Context(),w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
			return
		}

		userActions := user.NewActions(s.db)
		userModel, err := userActions.GetUserBySessionToken(r.Context(), token)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) || AppEnv.IsProd() {
				err = fmt.Errorf("unauthorized")
			}

			WriteError(r.Context(),w, http.StatusUnauthorized, err)
			return
		}

		r = r.WithContext(user.ContextWithUser(r.Context(), userModel))
		h.ServeHTTP(w, r)
	})
}

func (s *Server) LoggerInjector(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(s.log.WithContext(r.Context())))
	})
}
