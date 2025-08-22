package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/appenv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth"
	"github.com/rs/zerolog"
)

func Auth(authRepo auth.Repository) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Skip re-injecting Auth middleware if user/session info already exists in the context
			_, ok := auth.UserAndSessionFromContext(ctx)
			if ok {
				next.ServeHTTP(w, r)
				return
			}

			zlog := zerolog.Ctx(ctx)

			sendUnauthorizedError := func() {
				writeError(ctx, w, r, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
			}

			authStr := r.Header.Get("Authorization")
			if len(authStr) == 0 {
				authorizationCookie, authorizationCookieErr := r.Cookie("Authorization")
				if authorizationCookieErr != nil {
					sendUnauthorizedError()
					return
				}
				authStr = authorizationCookie.Value
			}

			token := strings.TrimSpace(strings.Replace(authStr, "Bearer", "", 1))
			if token == "" {
				sendUnauthorizedError()
				return
			}

			if _, err := authRepo.VerifyAuthToken(token); err != nil {
				if appenv.IsStagOrLocal() {
					zlog.Error().Err(err).Msg("Error from jwt verify function")
				}
				sendUnauthorizedError()
				return
			}

			userAndSessionData, err := authRepo.GetUserAndSessionDataBySessionToken(ctx, token)

			if err != nil {
				if errors.Is(err, apperr.ErrNoResult) {
					err = fmt.Errorf("unauthorized")
				} else if appenv.IsProd() {
					zlog.Error().Err(err).Msg("Error while geting a user by session tokne in auth middleware")
					err = fmt.Errorf("unauthorized")
				}

				writeError(ctx, w, r, http.StatusUnauthorized, err)
				return
			}

			// check blocking status
			blockedAt := userAndSessionData.UserBlockedAt
			blockedUntil := userAndSessionData.UserBlockedUntil
			if blockedAt.Valid || (blockedUntil.Valid && blockedUntil.Time.After(time.Now())) {
				writeError(ctx, w, r, http.StatusUnauthorized, apperr.ErrBlockedUser)
				return
			}

			ctx = auth.ContextWithUserAndSession(ctx, userAndSessionData)
			ctx = zlog.With().Int32("user_id", userAndSessionData.UserID).Int32("session_id", userAndSessionData.SessionID).Logger().WithContext(ctx)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

func (s *Server) LoggerInjector(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(s.zlog.WithContext(r.Context())))
	})
}

// Inject the Installation into the request context,
// will respond with error if it can't find the Installation in the database or in the request headers
func Installation(authRepo auth.Repository) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Skip re-injecting Installation middleware if the Installation info already exists in the context
			_, ok := auth.InstallationFromContext(ctx)
			if ok {
				next.ServeHTTP(w, r)
				return
			}

			zlog := zerolog.Ctx(ctx)
			missingOrInvalidErr := errors.New("missing A-Installation in the request header, or the the installation id is invalid")
			sendError := func() {
				writeError(ctx, w, r, http.StatusBadRequest, missingOrInvalidErr)
			}

			installationToken := r.Header.Get("A-Installation")
			if len(installationToken) == 0 {
				installationCookie, installationCookieErr := r.Cookie("A-Installation")
				if installationCookieErr != nil {
					sendError()
					return
				}
				installationToken = installationCookie.Value
			}
			if len(installationToken) == 0 {
				sendError()
				return
			}

			if _, err := authRepo.VerifyTokenForInstallation(installationToken); err != nil {
				if appenv.IsStagOrLocal() {
					zlog.Error().Err(err).Msg("Error from jwt verify function")
				}
				if errors.Is(err, apperr.ErrExpiredSessionToken) {
					writeError(ctx, w, r, http.StatusBadRequest, apperr.ErrExpiredInstallationSessionToken)
				} else {
					sendError()
				}
				return
			}

			var attachedToSessionId *int32
			userAndSession, ok := auth.UserAndSessionFromContext(ctx)
			if ok {
				attachedToSessionId = &userAndSession.SessionID
			}

			installation, err := authRepo.GetInstallationUsingToken(ctx, installationToken, attachedToSessionId)

			if err != nil {
				if errors.Is(err, apperr.ErrNoResult) {
					sendError()
					return
				}
				zlog.Error().Err(err).Msg("Error while geting a Installation from database in Installation middleware")
				if appenv.IsProd() {
					err = missingOrInvalidErr
				}
				writeError(ctx, w, r, http.StatusBadRequest, err)
				return
			}

			ctx = auth.ContextWithInstallation(ctx, installation)
			ctx = zlog.With().Int32("installation_id", installation.ID).Logger().WithContext(ctx)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}
