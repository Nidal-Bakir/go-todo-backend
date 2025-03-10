package apiutils

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Nidal-Bakir/go-todo-backend/internal/appenv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"

	"github.com/Nidal-Bakir/go-todo-backend/internal/tracker"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/mimes"
	"github.com/rs/zerolog"
)

type errorRes struct {
	Error  error   `json:"error"`
	Errors []error `json:"errors,omitempty"`
}

func (e errorRes) MarshalJSON() ([]byte, error) {
	m := make(map[string]any)
	m["error"] = e.Error.Error()

	errsLen := len(e.Errors)
	if errsLen != 0 {
		errors := make([]string, errsLen)
		for i, e := range e.Errors {
			errors[i] = e.Error()
		}
		m["errors"] = errors
	}

	return json.Marshal(m)
}

func WriteError(ctx context.Context, w http.ResponseWriter, code int, errs ...error) {
	// Assert(len(errs) != 0, "no errors to send")

	for i, e := range errs {
		appError := new(apperr.AppErr)
		ok := errors.As(e, appError)
		if ok {
			appError.SetTranslation(ctx)
			errs[i] = appError
		}
	}

	err := errorRes{Error: errs[0], Errors: errs[1:]}

	_writeJson(ctx, w, code, err, true)
}

func WriteJson(ctx context.Context, w http.ResponseWriter, code int, payload any) {
	_writeJson(ctx, w, code, payload, appenv.IsStagOrLocal())
}

func _writeJson(ctx context.Context, w http.ResponseWriter, code int, payload any, shouldLog bool) {
	zlog := *zerolog.Ctx(ctx)

	bytes, err := json.Marshal(payload)
	if err != nil {
		zlog.Error().Err(err).Any("payload", payload).Int("code", code).Msg("can not marshal payload in WriteJson")
		WriteError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Add("Content-Type", mimes.App_json)
	w.WriteHeader(code)
	w.Write(bytes)

	if shouldLog {
		_logRes(ctx, code, payload, zlog)
	}
}

func _logRes(ctx context.Context, code int, payload any, zerolog zerolog.Logger) {
	logEvent := zerolog.Info().Any("payload", payload).Int("code", code)
	reqId, ok := tracker.ReqUUIDFromContext(ctx)
	if ok {
		logEvent.Str(tracker.ReqIdStrKey, reqId.String())
	}
	logEvent.CallerSkipFrame(99999999) // so it dose not print the file:line_num in the log. we do not need those
	logEvent.Msg("Res")
}
