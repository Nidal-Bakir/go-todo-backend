package apperr

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Nidal-Bakir/go-todo-backend/internal/l10n"
)

type AppErr struct {
	err           error
	translationID string
	translatedMsg string
	errorCode     string
}

func (err AppErr) Error() string {
	if err.translatedMsg != "" {
		return err.translatedMsg
	}
	return err.err.Error()
}

func (err AppErr) Unwrap() error { return err.err }

func (err AppErr) ErrorCode() string { return err.errorCode }

func (err *AppErr) SetTranslation(ctx context.Context) {
	local, ok := l10n.LocalizerFromContext(ctx)
	if ok && err.translationID != "" {
		err.translatedMsg = local.GetWithId(err.translationID)
	}
}

func (e AppErr) MarshalJSON() ([]byte, error) {
	m := make(map[string]any, 2)

	m["error"] = e.Error()

	if len(e.ErrorCode()) != 0 {
		m["code"] = e.ErrorCode()
	}

	return json.Marshal(m)
}

func NewAppErr(err error) error {
	return &AppErr{
		err: err,
	}
}

func NewAppErrWithErrorCode(err error, errorCode string) error {
	return &AppErr{
		err:       err,
		errorCode: errorCode,
	}
}

func NewAppErrWithTr(err error, translationID string, errorCode string) error {
	return &AppErr{
		err:           err,
		translationID: translationID,
		errorCode:     errorCode,
	}
}

// -------------------------------------------

var (
	ErrNoResult = NewAppErrWithTr(errors.New("error no result found"), "no_result_found", "res_1")

	ErrUnexpectedErrorOccurred = NewAppErrWithTr(errors.New("unexpected error occurred"), "unexpected_error_occurred", "res_2")

	ErrTooManyRequests = NewAppErrWithTr(errors.New("too many requests"), "too_many_requests", "res_3")

	// auth
	ErrInvalidEmail            = NewAppErrWithTr(errors.New("invalid email"), "invalid_email", "auth_1")
	ErrInvalidPhoneNumber      = NewAppErrWithTr(errors.New("invalid phone number"), "invalid_phone_number", "auth_2")
	ErrUnsupportedLoginMethod  = NewAppErrWithTr(errors.New("unsupported login method"), "unsupported_login_method", "auth_3")
	ErrTooShortPassword        = NewAppErrWithTr(errors.New("too short password"), "too_short_password", "auth_4")
	ErrTooShortName            = NewAppErrWithTr(errors.New("too short name"), "too_short_name", "auth_5")
	ErrInvalidOtpCode          = NewAppErrWithTr(errors.New("error invalid otp code"), "invalid_otp_code", "auth_6")
	ErrInvalidTempUserdata     = NewAppErrWithTr(errors.New("error invalid temp user data"), "invalid_data", "auth_7")
	ErrInvalidLoginCredentials = NewAppErrWithTr(errors.New("error invalid login credentials"), "invalid_login_credentials", "auth_8")
	ErrAlreadyUsedEmail        = NewAppErrWithTr(errors.New("error already used email"), "already_used_email", "auth_10")
	ErrAlreadyUsedPhoneNumber  = NewAppErrWithTr(errors.New("error already used phone number"), "already_used_phone_number", "auth_11")
)
