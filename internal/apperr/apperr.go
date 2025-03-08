package apperr

import "errors"

type AppErr struct {
	err           error
	translationID string
}

func (err *AppErr) Error() string { return err.err.Error() }
func (err *AppErr) Unwrap() error { return err.err }

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
	ErrNoResult = NewAppErrWithTr(errors.New("error no result found"), "no_result_found")
)
