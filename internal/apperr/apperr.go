package apperr

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Nidal-Bakir/go-todo-backend/internal/l10n"
)

func IsAppErr(err error) bool {
	return UnwrapAppErr(err) != nil
}

func UnwrapAppErr(err error) *AppErr {
	for {
		appErr, ok := err.(*AppErr)
		if ok {
			return appErr
		}
		switch x := err.(type) {
		case interface{ Unwrap() error }:
			err = x.Unwrap()
			if err == nil {
				return nil
			}
		case interface{ Unwrap() []error }:
			for _, err := range x.Unwrap() {
				e := UnwrapAppErr(err)
				if e != nil {
					return e
				}
			}
			return nil
		default:
			return nil
		}
	}
}

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
	ErrNoResult                = NewAppErrWithTr(errors.New("error no result found"), l10n.NoResultFoundTrId, "res_1")
	ErrUnexpectedErrorOccurred = NewAppErrWithTr(errors.New("unexpected error occurred"), l10n.UnexpectedErrorOccurredTrId, "res_2")
	ErrTooManyRequests         = NewAppErrWithTr(errors.New("too many requests"), l10n.TooManyRequestsTrId, "res_3")
	ErrInvalidId               = NewAppErrWithTr(errors.New("invalid id"), l10n.InvalidId, "res_4")

	// auth
	ErrInvalidEmail                      = NewAppErrWithTr(errors.New("invalid email"), l10n.InvalidEmailTrId, "auth_1")
	ErrInvalidPhoneNumber                = NewAppErrWithTr(errors.New("invalid phone number"), l10n.InvalidPhoneNumberTrId, "auth_2")
	ErrUnsupportedLoginMethod            = NewAppErrWithTr(errors.New("unsupported login method"), l10n.UnsupportedLoginMethodTrId, "auth_3")
	ErrTooShortPassword                  = NewAppErrWithTr(errors.New("too short password"), l10n.TooShortPasswordTrId, "auth_4")
	ErrTooShortName                      = NewAppErrWithTr(errors.New("too short name"), l10n.TooShortNameTrId, "auth_5")
	ErrInvalidOtpCode                    = NewAppErrWithTr(errors.New("invalid otp code"), l10n.InvalidOtpCodeTrId, "auth_6")
	ErrInvalidTempUserdata               = NewAppErrWithTr(errors.New("invalid temp user data"), l10n.InvalidDataTrId, "auth_7")
	ErrInvalidLoginCredentials           = NewAppErrWithTr(errors.New("invalid login credentials"), l10n.InvalidLoginCredentialsTrId, "auth_8")
	ErrAlreadyUsedEmail                  = NewAppErrWithTr(errors.New("already used email"), l10n.AlreadyUsedEmailTrId, "auth_10")
	ErrAlreadyUsedPhoneNumber            = NewAppErrWithTr(errors.New("already used phone number"), l10n.AlreadyUsedPhoneNumberTrId, "auth_11")
	ErrOldPasswordDoesNotMatchCurrentOne = NewAppErrWithTr(errors.New("old password noes not match current one"), l10n.OldPasswordDoesNotMatchCurrentOneTrId, "auth_12")
	ErrInstallationTokenInUse            = NewAppErrWithErrorCode(errors.New("cannot link with the provided installation token — it is already linked to another user, or the current user if you did not unlinked(logout) yet"), "auth_13")

	// jwt
	ErrExpiredSessionToken             = NewAppErrWithTr(errors.New("expired session token"), l10n.ExpiredSessionToken, "auth_13")
	ErrExpiredInstallationSessionToken = NewAppErrWithErrorCode(errors.New("expired installation session token"), "auth_14")

	// user
	ErrBlockedUser = NewAppErrWithTr(errors.New("blocked user"), l10n.BlockedUser, "user_1")

	// todo
	ErrUnsupportedTodoStatus = NewAppErrWithTr(errors.New("unsupported todo status"), l10n.UnsupportedTodoStatus, "todo_1")
)
