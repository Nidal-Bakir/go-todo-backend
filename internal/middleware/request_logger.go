package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func RequestLogger(next http.Handler) http.HandlerFunc {
	return hlog.AccessHandler(func(r *http.Request, status int, size int, duration time.Duration) {
		infoLog := zerolog.Ctx(r.Context()).Info()

		infoLog.
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			CallerSkipFrame(99999999) // so it dose not print the file:line_num in the log. we do not need those

		headersSubDict := zerolog.Dict()
		for k, v := range r.Header {
			headersSubDict.Str(k, strings.Join(v, ";"))
		}
		infoLog.Dict("headers", headersSubDict)

		infoLog.Msg("Req")
	})(next).ServeHTTP
}