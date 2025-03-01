package user

import (
	"errors"

	apperr "github.com/Nidal-Bakir/go-todo-backend/internal/app_error"
)

var (
	ErrInvalidOtpCode          = apperr.NewAppErrWithTr(errors.New("error invalid otp code"), "invalid_otp_code")
	ErrInvalidTempUserdata     = apperr.NewAppErrWithTr(errors.New("error invalid temp user data"), "invalid_data")
	ErrInvalidLoginCredentials = apperr.NewAppErrWithTr(errors.New("error invalid login credentials"), "invalid_login_credentials")
)
