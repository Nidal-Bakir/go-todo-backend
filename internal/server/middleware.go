package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Nidal-Bakir/go-todo-backend/internal/appenv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/user"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/appjwt"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/password_hasher"
	"github.com/rs/zerolog"
)

func (s *Server) Auth(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zlog := zerolog.Ctx(r.Context())

		token := strings.TrimSpace(strings.Replace(r.Header.Get("Authorization"), "Bearer", "", 1))
		if token == "" {
			WriteError(r.Context(), w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
			return
		}

		if _, err := appjwt.NewAppJWT().VerifyToken(token, "auth"); err != nil {
			if appenv.IsStagOrLocal() {
				zlog.Error().Err(err).Msg("Error from jwt verify function")
			}
			WriteError(r.Context(), w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
			return
		}

		userRepo := user.NewRepository(
			user.NewDataSource(s.db, s.rdb),
			s.gatewaysProvider,
			password_hasher.NewPasswordHasher(password_hasher.BcryptPasswordHash),
		)
		userModel, err := userRepo.GetUserBySessionToken(r.Context(), token)

		if err != nil {
			if errors.Is(err, apperr.ErrNoResult) {
				err = fmt.Errorf("unauthorized")
			} else if appenv.IsProd() {
				zlog.Error().Err(err).Msg("Error while geting a user by session tokne in auth middleware")
				err = fmt.Errorf("unauthorized")
			}

			WriteError(r.Context(), w, http.StatusUnauthorized, err)
			return
		}

		r = r.WithContext(user.ContextWithUser(r.Context(), userModel))
		h.ServeHTTP(w, r)
	})
}

func (s *Server) LoggerInjector(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(s.zlog.WithContext(r.Context())))
	})
}
