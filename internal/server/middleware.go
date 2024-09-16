package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/AppEnv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/l10n"
	"github.com/Nidal-Bakir/go-todo-backend/internal/tracker"
	"github.com/Nidal-Bakir/go-todo-backend/internal/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/hlog"
)

type Middleware func(http.Handler) http.HandlerFunc

func MiddlewareChain(h http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	if len(m) < 1 {
		return h
	}

	wrapped := h

	// loop in reverse to preserve middleware order
	for i := len(m) - 1; i >= 0; i-- {
		wrapped = m[i](wrapped)
	}

	return wrapped
}

func (s *Server) Auth(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.Replace(r.Header.Get("Authorization"), "Bearer", "", 1))
		// TODO: also check if its a valid token (length, schema etc..)
		if token == "" {
			WriteError(r.Context(), w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
			return
		}

		userActions := user.NewRepository(s.db)
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

func (s *Server) RequestLogger(h http.Handler) http.HandlerFunc {
	return hlog.AccessHandler(func(r *http.Request, status int, size int, duration time.Duration) {
		logInfo := s.log.Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			CallerSkipFrame(99999) // so it dose not print the file:line_num in the log. we do not need those

		if id, ok := tracker.ReqUUIDFromContext(r.Context()); ok {
			logInfo.Str("req_uuid", id.String())
		}

		logInfo.Send()
	})(h).ServeHTTP
}

func (s *Server) RequestUUIDMiddleware(h http.Handler) http.HandlerFunc {
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

func (s *Server) LoggerInjector(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(s.log.WithContext(r.Context())))
	})
}

func (s *Server) LocalizerInjector(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := r.Header.Get("Accept-Language")
		langQP := r.FormValue("lang")
		if langQP != "" {
			lang = langQP
		}

		if lang == "" {
			WriteError(r.Context(), w, http.StatusBadRequest, errors.New("missing Accept-Language in Headers or lang in Query Parameter"))
			return
		}

		h.ServeHTTP(w, r.WithContext(l10n.ContextWithLocalizer(r.Context(), l10n.GetLocalizer(lang))))
	})
}

func (s *Server) Heartbeat(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (r.Method == "GET" || r.Method == "HEAD") && strings.EqualFold(r.URL.Path, "/ping") {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("pong"))
			return
		}
		h.ServeHTTP(w, r)
	})
}
