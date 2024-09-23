package middleware

import (
	"net/http"

	"github.com/Nidal-Bakir/go-todo-backend/internal/tracker"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func RequestUUIDMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := *zerolog.Ctx(r.Context())

		var uuidVal uuid.UUID
		uuidStr := r.Header.Get("X-Request-UUID")
		if uuidStr == "" {
			uuidVal = uuid.New()
		} else {
			if u, err := uuid.Parse(uuidStr); err == nil {
				uuidVal = u
			} else {
				log.
					Error().
					Err(err).
					Str("uuidStr", uuidStr).
					Str("x-header", "X-Request-UUID").
					Msg("Error while parsing a uuid from request header")
				uuidVal = uuid.New()
			}
		}

		ctx = tracker.ContextWithReqUUID(ctx, uuidVal)
		ctx = log.With().Str(tracker.ReqIdStrKey, uuidVal.String()).Logger().WithContext(ctx)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
