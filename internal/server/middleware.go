package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/AppEnv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/tracker"
	"github.com/Nidal-Bakir/go-todo-backend/internal/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/hlog"
)

func (s *Server) Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.Replace(r.Header.Get("Authorization"), "Bearer", "", 1))
		// TODO: also check if its a valid token (length, schema etc..)
		if token == "" {
			WriteError(r.Context(), w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
			return
		}

		userActions := user.NewActions(s.db)
		userModel, err := userActions.GetUserBySessionToken(r.Context(), token)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				err = fmt.Errorf("unauthorized")
			} else if AppEnv.IsProd() {
				s.log.Error().Err(err).Msg("Error while geting a user by session tokne in auth middleware")
				err = fmt.Errorf("unauthorized")
			}

			WriteError(r.Context(), w, http.StatusUnauthorized, err)
			return
		}

		r = r.WithContext(user.ContextWithUser(r.Context(), userModel))
		h.ServeHTTP(w, r)
	})
}

func (s *Server) RequsetLogger(h http.Handler) http.Handler {
	return hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		logInfo := s.log.Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			CallerSkipFrame(99999)

		if id, ok := tracker.ReqUUIDFromContext(r.Context()); ok {
			logInfo.Str("req_uuid", id.String())
		}

		logInfo.Send()
	})(h)
}

func (s *Server) RequestUUIDMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var uuidVal uuid.UUID

		uuidStr := r.Header.Get("X-Request-UUID")
		if uuidStr == "" {
			uuidVal = uuid.New()
		} else {
			if u, err := uuid.Parse(uuidStr); err == nil {
				uuidVal = u
			} else {
				s.log.Error().Err(err).Str("uuidStr", uuidStr).Str("x-header", "X-Request-UUID").Msg("Error while parsing a uuid from request header")
				uuidVal = uuid.New()
			}
		}

		ctx := tracker.ContextWithReqUUID(r.Context(), uuidVal)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) LoggerInjector(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(s.log.WithContext(r.Context())))
	})
}
