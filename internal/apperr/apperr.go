package apperr

import (
	"context"
	"errors"

	"github.com/Nidal-Bakir/go-todo-backend/internal/l10n"
)

type AppErr struct {
	err           error
	translationID string
	translatedMsg string
}

func (err AppErr) Error() string {
	if err.translatedMsg != "" {
		return err.translatedMsg
	}
	return err.err.Error()
}
func (err AppErr) Unwrap() error { return err.err }

func (err *AppErr) SetTranslation(ctx context.Context) {
	local, ok := l10n.LocalizerFromContext(ctx)
	if ok && err.translationID != "" {
		err.translatedMsg = local.GetWithId(err.translationID)
	}
}

func NewAppErr(err error) error {
	return &AppErr{
		err: err,
	}
}

func NewAppErrWithTr(err error, translationID string) error {
	return &AppErr{
		err:           err,
		translationID: translationID,
	}
}

// -------------------------------------------

var (
	ErrNoResult                = NewAppErrWithTr(errors.New("error no result found"), "no_result_found")
	ErrUnexpectedErrorOccurred = NewAppErrWithTr(errors.New("unexpected error occurred"), "unexpected_error_occurred")
	ErrTooManyRequests         = NewAppErrWithTr(errors.New("too many requests"), "too_many_requests")
	ErrInvalidEmail            = NewAppErrWithTr(errors.New("invalid email"), "invalid_email")
)
