package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Nidal-Bakir/go-todo-backend/internal/AppEnv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/tracker"
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
	zlog := *zerolog.Ctx(ctx)

	var err errorRes
	if len(errs) == 0 {
		if AppEnv.IsStagOrLocal() {
			panic("WriteError: empty errs array")
		}
		zlog.Warn().Msg("WriteError: empty errs array")
		err = errorRes{Error: fmt.Errorf("empty errs array"), Errors: []error{}}
	} else {
		err = errorRes{Error: errs[0], Errors: errs[1:]}
	}

	_writeJson(ctx, w, code, err, true)
}

func WriteJson(ctx context.Context, w http.ResponseWriter, code int, payload any) {
	_writeJson(ctx, w, code, payload, AppEnv.IsStagOrLocal())
}

func _writeJson(ctx context.Context, w http.ResponseWriter, code int, payload any, shouldLog bool) {
	zlog := *zerolog.Ctx(ctx)

	bytes, err := json.Marshal(payload)
	if err != nil {
		zlog.Error().Err(err).Any("payload", payload).Int("code", code).Msg("can not marshal payload in WriteJson")
		WriteError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
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
