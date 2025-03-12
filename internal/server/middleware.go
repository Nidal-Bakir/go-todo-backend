package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Nidal-Bakir/go-todo-backend/internal/appenv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/appjwt"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func (s *Server) Auth(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		zlog := zerolog.Ctx(ctx)

		token := strings.TrimSpace(strings.Replace(r.Header.Get("Authorization"), "Bearer", "", 1))
		if token == "" {
			WriteError(ctx, w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
			return
		}

		if _, err := appjwt.NewAppJWT().VerifyToken(token, "auth"); err != nil {
			if appenv.IsStagOrLocal() {
				zlog.Error().Err(err).Msg("Error from jwt verify function")
			}
			WriteError(ctx, w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
			return
		}

		authRepo := s.NewAuthRepository()
		userModel, err := authRepo.GetUserBySessionToken(ctx, token)

		if err != nil {
			if errors.Is(err, apperr.ErrNoResult) {
				err = fmt.Errorf("unauthorized")
			} else if appenv.IsProd() {
				zlog.Error().Err(err).Msg("Error while geting a user by session tokne in auth middleware")
				err = fmt.Errorf("unauthorized")
			}

			WriteError(ctx, w, http.StatusUnauthorized, err)
			return
		}

		r = r.WithContext(auth.ContextWithUser(ctx, userModel))
		next.ServeHTTP(w, r)
	})
}

func (s *Server) LoggerInjector(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(s.zlog.WithContext(r.Context())))
	})
}

// Inject the Installation into the request context,
// will respond with error if it can't find the Installation in the database
func (s *Server) Installation(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		zlog := zerolog.Ctx(ctx)

		missingOrInvalidErr := errors.New("missing A-Installation in the request header, or the the installation id is invalid")
		sendError := func() {
			WriteError(ctx, w, http.StatusBadRequest, missingOrInvalidErr)
		}

		installationId, err := uuid.Parse(r.Header.Get("A-Installation"))
		if err != nil {
			sendError()
			return
		}

		var attachedToUserId *int32
		user, ok := auth.UserFromContext(ctx)
		if ok {
			attachedToUserId = &user.ID
		}

		authRepo := s.NewAuthRepository()
		installation, err := authRepo.GetInstallationUsingUuid(ctx, installationId, attachedToUserId)

		if err != nil {
			if errors.Is(err, apperr.ErrNoResult) {
				sendError()
				return
			}
			zlog.Error().Err(err).Msg("Error while geting a Installation from database in Installation middleware")
			if appenv.IsProd() {
				err = missingOrInvalidErr
			}
			WriteError(ctx, w, http.StatusBadRequest, err)
			return
		}

		r = r.WithContext(auth.ContextWithInstallation(ctx, installation))
		next.ServeHTTP(w, r)
	})
}
