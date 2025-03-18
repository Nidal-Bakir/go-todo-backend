package apiutils

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Nidal-Bakir/go-todo-backend/internal/appenv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/mimes"
	"github.com/rs/zerolog"
)

type errorRes struct {
	Errors []error `json:"errors"`
}

func (e errorRes) MarshalJSON() ([]byte, error) {
	m := make(map[string]any)

	errorList := make([]any, len(e.Errors))
	for i, e := range e.Errors {
		if _, ok := e.(*apperr.AppErr); ok {
			errorList[i] = e
		} else {
			errorList[i] = map[string]string{"error": e.Error()}
		}
	}
	m["errors"] = errorList

	return json.Marshal(m)
}

func WriteError(ctx context.Context, w http.ResponseWriter, code int, errs ...error) {
	for i, e := range errs {
		appError := new(apperr.AppErr)
		ok := errors.As(e, appError)
		if ok {
			appError.SetTranslation(ctx)
			errs[i] = appError
		}
	}
	err := errorRes{Errors: errs}

	// on production log the actual error, and send an arbitrary error to the user
	if code == http.StatusInternalServerError && appenv.IsProd() {
		_logRes(code, err, *zerolog.Ctx(ctx))
		err := apperr.ErrUnexpectedErrorOccurred.(apperr.AppErr)
		err.SetTranslation(ctx)
		_writeJson(ctx, w, code, err, false)
		return
	}

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
		_logRes(code, payload, zlog)
	}
}

func _logRes(code int, payload any, zerolog zerolog.Logger) {
	logEvent := zerolog.Info().Any("payload", payload).Int("code", code)
	logEvent.CallerSkipFrame(99999999) // so it dose not print the file:line_num in the log. we do not need those
	logEvent.Msg("Res")
}
