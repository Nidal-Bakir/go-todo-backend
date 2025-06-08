package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter/redis_ratelimiter"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/emailvalidator"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

func authRouter(ctx context.Context, s *Server, authRepo auth.Repository) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"POST /create-account",
		middleware.MiddlewareChain(
			createTempAccount(authRepo),
			middleware.ACT_app_x_www_form_urlencoded,
			createAcccountRateLimiterByIP(ctx, s.rdb),
			createAcccountRateLimiterByAccessKey(ctx, s.rdb),
		),
	)

	mux.HandleFunc(
		"POST /verify-account",
		middleware.MiddlewareChain(
			vareifyAccount(authRepo),
			middleware.ACT_app_x_www_form_urlencoded,
			verifyAccountRateLimiterById(ctx, s.rdb),
		),
	)

	mux.HandleFunc(
		"POST /login",
		middleware.MiddlewareChain(
			login(authRepo),
			middleware.ACT_app_x_www_form_urlencoded,
			Installation(authRepo),
			loginRateLimiter(ctx, s.rdb),
		),
	)

	mux.HandleFunc(
		"POST /logout",
		middleware.MiddlewareChain(
			logout(authRepo),
			middleware.ACT_app_x_www_form_urlencoded,
			Auth(authRepo),
			Installation(authRepo),
		),
	)

	// for logged-in user
	mux.HandleFunc(
		"POST /change-password",
		middleware.MiddlewareChain(
			changeUserPassword(authRepo),
			middleware.ACT_app_x_www_form_urlencoded,
			Auth(authRepo),
		),
	)

	mux.HandleFunc(
		"POST /forget-password",
		middleware.MiddlewareChain(
			forgetPassword(authRepo),
			middleware.ACT_app_x_www_form_urlencoded,
			forgetPasswordRateLimiterByIP(ctx, s.rdb),
			forgetPasswordRateLimiterByAccessKey(ctx, s.rdb),
		),
	)
	mux.HandleFunc(
		"POST /reset-password",
		middleware.MiddlewareChain(
			resetPassword(authRepo),
			middleware.ACT_app_x_www_form_urlencoded,
			resetPasswordRateLimiterById(ctx, s.rdb),
		),
	)

	return middleware.MiddlewareChain(
		mux.ServeHTTP,
	)
}

//-----------------------------------------------------------------------------

func createAcccountRateLimiterByIP(ctx context.Context, rdb *redis.Client) func(next http.Handler) http.HandlerFunc {
	return middleware.RateLimiter(
		func(r *http.Request) (string, error) {
			return r.RemoteAddr, nil
		},
		redis_ratelimiter.NewRedisSlidingWindowLimiter(
			ctx,
			rdb,
			ratelimiter.Config{

				PerTimeFrame: 100,
				TimeFrame:    time.Hour * 24,
				KeyPrefix:    "auth:create:account:ip",
			},
		),
	)
}

func createAcccountRateLimiterByAccessKey(ctx context.Context, rdb *redis.Client) func(next http.Handler) http.HandlerFunc {
	return middleware.RateLimiter(
		func(r *http.Request) (string, error) {
			err := r.ParseForm()
			if err != nil {
				return "", err
			}

			createAccountParam, _ := validateCreateAccountParam(r)

			key := ""
			createAccountParam.LoginMethod.FoldOr(
				func() {
					key = createAccountParam.Email
				},
				func() {
					key = createAccountParam.PhoneNumber.ToAppStanderdForm()
				},
				func() {
					// Even if the validation has errors, do not return an error.
					// Instead, set the key to the string "unknown_login_method"
					// and allow the misbehaving user to be rate-limited along with other
					// misbehaving users.
					key = "unknown_login_method"
				},
			)
			return key, nil
		},
		redis_ratelimiter.NewRedisSlidingWindowLimiter(
			ctx,
			rdb,
			ratelimiter.Config{

				PerTimeFrame: 20,
				TimeFrame:    time.Hour * 24,
				KeyPrefix:    "auth:create:account:access_key",
			},
		),
	)
}

func loginRateLimiter(ctx context.Context, rdb *redis.Client) func(next http.Handler) http.HandlerFunc {
	return middleware.RateLimiter(
		func(r *http.Request) (string, error) {
			err := r.ParseForm()
			if err != nil {
				return "", err
			}

			loginParam, _ := validateLoginParam(r)

			key := ""
			loginParam.LoginMethod.FoldOr(
				func() {
					key = loginParam.Email
				},
				func() {
					key = loginParam.PhoneNumber.ToAppStanderdForm()
				},
				func() {
					// Even if the validation has errors, do not return an error.
					// Instead, set the key to the string "unknown_login_method"
					// and allow the misbehaving user to be rate-limited along with other
					// misbehaving users.
					key = "unknown_login_method"
				},
			)
			return key, nil
		},
		redis_ratelimiter.NewRedisSlidingWindowLimiter(
			ctx,
			rdb,
			ratelimiter.Config{
				PerTimeFrame: 10,
				TimeFrame:    time.Hour * 12,
				KeyPrefix:    "auth:login",
			},
		),
	)
}

func verifyAccountRateLimiterById(ctx context.Context, rdb *redis.Client) func(next http.Handler) http.HandlerFunc {
	return middleware.RateLimiter(
		func(r *http.Request) (string, error) {
			err := r.ParseForm()
			if err != nil {
				return "", err
			}
			param, _ := validateVareifyAccountParams(r)
			return param.Id.String(), nil
		},
		redis_ratelimiter.NewRedisSlidingWindowLimiter(
			ctx,
			rdb,
			ratelimiter.Config{
				PerTimeFrame: 5,
				TimeFrame:    time.Minute * 15,
				KeyPrefix:    "auth:verify:account:id",
			},
		),
	)
}

func resetPasswordRateLimiterById(ctx context.Context, rdb *redis.Client) func(next http.Handler) http.HandlerFunc {
	return middleware.RateLimiter(
		func(r *http.Request) (string, error) {
			err := r.ParseForm()
			if err != nil {
				return "", err
			}
			param, _ := validateResetPasswordParams(r)
			return param.Id.String(), nil
		},
		redis_ratelimiter.NewRedisSlidingWindowLimiter(
			ctx,
			rdb,
			ratelimiter.Config{
				PerTimeFrame: 5,
				TimeFrame:    time.Minute * 15,
				KeyPrefix:    "auth:reset:password:id",
			},
		),
	)
}

func forgetPasswordRateLimiterByIP(ctx context.Context, rdb *redis.Client) func(next http.Handler) http.HandlerFunc {
	return middleware.RateLimiter(
		func(r *http.Request) (string, error) {
			return r.RemoteAddr, nil
		},
		redis_ratelimiter.NewRedisSlidingWindowLimiter(
			ctx,
			rdb,
			ratelimiter.Config{
				PerTimeFrame: 100,
				TimeFrame:    time.Hour * 24,
				KeyPrefix:    "auth:forget:password:ip",
			},
		),
	)
}

func forgetPasswordRateLimiterByAccessKey(ctx context.Context, rdb *redis.Client) func(next http.Handler) http.HandlerFunc {
	return middleware.RateLimiter(
		func(r *http.Request) (string, error) {
			err := r.ParseForm()
			if err != nil {
				return "", err
			}

			param, _ := validateForgetPasswordParam(r)

			key := ""
			param.LoginMethod.FoldOr(
				func() {
					key = param.Email
				},
				func() {
					key = param.PhoneNumber.ToAppStanderdForm()
				},
				func() {
					// Even if the validation has errors, do not return an error.
					// Instead, set the key to the string "unknown_login_method"
					// and allow the misbehaving user to be rate-limited along with other
					// misbehaving users.
					key = "unknown_login_method"
				},
			)
			return key, nil
		},
		redis_ratelimiter.NewRedisSlidingWindowLimiter(
			ctx,
			rdb,
			ratelimiter.Config{

				PerTimeFrame: 10,
				TimeFrame:    time.Hour * 24,
				KeyPrefix:    "auth:forget:password:access_key",
			},
		),
	)
}

//-----------------------------------------------------------------------------

type createAccountParams struct {
	LoginMethod auth.LoginMethod
	Email       string
	PhoneNumber utils.PhoneNumber
	Password    string
	FirstName   string
	LastName    string
}

func createTempAccount(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, http.StatusBadRequest, err)
			return
		}

		createAccountParam, errList := validateCreateAccountParam(r)
		if len(errList) != 0 {
			writeError(ctx, w, http.StatusBadRequest, errList...)
			return
		}

		tuser := new(auth.TempUser)
		tuser.LoginMethod = createAccountParam.LoginMethod
		tuser.Email = createAccountParam.Email
		tuser.Phone = createAccountParam.PhoneNumber
		tuser.Password = createAccountParam.Password
		tuser.Lname = createAccountParam.LastName
		tuser.Fname = createAccountParam.FirstName

		tuser, err = authRepo.CreateTempUser(ctx, tuser)
		if err != nil {
			writeError(ctx, w, return400IfAppErrOr500(err), err)
			return
		}

		response := struct {
			Id string `json:"id"`
		}{
			Id: tuser.Id.String(),
		}

		writeJson(ctx, w, http.StatusCreated, response)
	}
}

func validateCreateAccountParam(r *http.Request) (createAccountParams, []error) {
	loginMethodFormStr := r.FormValue("login_method")

	emailFormStr := r.FormValue("email")

	phone := utils.PhoneNumber{
		CountryCode: r.FormValue("country_code"),
		Number:      r.FormValue("phone_number"),
	}

	passwordFormStr := r.FormValue("password")

	fNameFormStr := r.FormValue("f_name")
	lNameFormStr := r.FormValue("l_name")

	errList := make([]error, 0, 5)

	loginMethod, err := new(auth.LoginMethod).FromString(loginMethodFormStr)
	if err != nil {
		errList = append(errList, err)
	}

	errList = append(errList, validatePassword(passwordFormStr)...)

	loginMethod.FoldOr(
		func() {
			if !emailvalidator.IsValidEmail(emailFormStr) {
				errList = append(errList, apperr.ErrInvalidEmail)
			}
		},
		func() {
			if !phone.IsValid() {
				errList = append(errList, apperr.ErrInvalidPhoneNumber)
			}
		},
		func() {
			// no op
		},
	)

	if len(fNameFormStr) <= 2 {
		errList = append(errList, apperr.ErrTooShortName)
	}
	if len(lNameFormStr) <= 2 {
		errList = append(errList, apperr.ErrTooShortName)
	}

	if len(errList) != 0 {
		return createAccountParams{}, errList
	}

	createAccountParam := createAccountParams{
		LoginMethod: *loginMethod,
		Email:       emailFormStr,
		PhoneNumber: phone,
		Password:    passwordFormStr,
		FirstName:   fNameFormStr,
		LastName:    lNameFormStr,
	}

	return createAccountParam, errList
}

//-----------------------------------------------------------------------------

type vareifyAccountParams struct {
	Id   uuid.UUID
	Code string
}

func vareifyAccount(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, http.StatusBadRequest, err)
			return
		}

		vareifyAccountParam, errList := validateVareifyAccountParams(r)
		if len(errList) != 0 {
			writeError(ctx, w, http.StatusBadRequest, errList...)
			return
		}

		user, err := authRepo.CreateUser(ctx, vareifyAccountParam.Id, vareifyAccountParam.Code)

		if err != nil {
			writeError(ctx, w, return400IfAppErrOr500(err), err)
			return
		}

		response := struct {
			User publicUser `json:"user"`
		}{
			User: NewPublicUserFromAuthUser(user),
		}

		writeJson(ctx, w, http.StatusCreated, response)
	}
}

func validateVareifyAccountParams(r *http.Request) (vareifyAccountParams, []error) {
	idFormStr := r.FormValue("id")
	code := r.FormValue("code")

	errList := make([]error, 0, 2)

	tUserUUID, err := uuid.Parse(idFormStr)
	if err != nil {
		errList = append(errList, errors.New("invalid id"))
	}

	if len(code) != auth.OtpCodeLength {
		errList = append(errList, errors.New("invalid code"))
	}

	if len(errList) != 0 {
		return vareifyAccountParams{}, errList
	}

	params := vareifyAccountParams{
		Id:   tUserUUID,
		Code: code,
	}

	return params, errList
}

//-----------------------------------------------------------------------------

type loginParams struct {
	LoginMethod auth.LoginMethod
	Email       string
	PhoneNumber utils.PhoneNumber
	Password    string
}

func login(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, http.StatusBadRequest, err)
			return
		}

		loginParam, errList := validateLoginParam(r)
		if len(errList) != 0 {
			writeError(ctx, w, http.StatusBadRequest, errList...)
			return
		}

		installation, ok := auth.InstallationFromContext(ctx)
		utils.Assert(ok, "we should find the installation in the context tree, but we did not. something is wrong.")

		user, token, err := authRepo.PasswordLogin(
			ctx,
			auth.PasswordLoginAccessKey{Phone: loginParam.PhoneNumber, Email: loginParam.Email, LoginMethod: loginParam.LoginMethod},
			loginParam.Password,
			installation,
		)
		if err != nil {
			statusCode := return400IfAppErrOr500(err)
			if errors.Is(err, apperr.ErrInvalidLoginCredentials) {
				statusCode = http.StatusUnauthorized
			}
			writeError(ctx, w, statusCode, err)
			return
		}

		response := struct {
			User  publicUser `json:"user"`
			Token string     `json:"token"`
		}{
			User:  NewPublicUserFromAuthUser(user),
			Token: token,
		}
		writeJson(ctx, w, http.StatusCreated, response)
	}
}

func validateLoginParam(r *http.Request) (loginParams, []error) {
	loginMethodFormStr := r.FormValue("login_method")

	emailFormStr := r.FormValue("email")

	phone := utils.PhoneNumber{
		CountryCode: r.FormValue("country_code"),
		Number:      r.FormValue("phone_number"),
	}

	passwordFormStr := r.FormValue("password")

	errList := make([]error, 0, 3)

	loginMethod, err := new(auth.LoginMethod).FromString(loginMethodFormStr)
	if err != nil {
		errList = append(errList, err)
	}

	loginMethod.FoldOr(
		func() {
			if !emailvalidator.IsValidEmail(emailFormStr) {
				errList = append(errList, apperr.ErrInvalidEmail)
			}
		},
		func() {
			if !phone.IsValid() {
				errList = append(errList, apperr.ErrInvalidPhoneNumber)
			}
		},
		func() {
			// no op
		},
	)

	errList = append(errList, validatePassword(passwordFormStr)...)

	if len(errList) != 0 {
		return loginParams{}, errList
	}

	loginParam := loginParams{
		LoginMethod: *loginMethod,
		Email:       emailFormStr,
		PhoneNumber: phone,
		Password:    passwordFormStr,
	}

	return loginParam, errList
}

//-----------------------------------------------------------------------------

type logoutParams struct {
	terminateAllOtherSessions bool
}

func logout(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, http.StatusBadRequest, err)
			return
		}

		logoutParam, errList := validateLogoutParam(r)
		if len(errList) != 0 {
			writeError(ctx, w, http.StatusBadRequest, errList...)
			return
		}

		userAndSession, ok := auth.UserAndSessionFromContext(ctx)
		utils.Assert(ok, "the user should be in the ctx")

		installation, ok := auth.InstallationFromContext(ctx)
		utils.Assert(ok, "the installation should be in the ctx")

		err = authRepo.Logout(
			ctx,
			int(userAndSession.UserID),
			int(installation.ID),
			int(userAndSession.SessionID),
			logoutParam.terminateAllOtherSessions,
		)
		if err != nil {
			writeError(ctx, w, return400IfAppErrOr500(err), err)
			return
		}

		writeOperationDoneSuccessfullyJson(ctx, w)
	}
}

func validateLogoutParam(r *http.Request) (logoutParams, []error) {
	params := logoutParams{}
	errList := make([]error, 0, 1)

	if t, err := strconv.ParseBool(r.FormValue("terminate_all_other_sessions")); err != nil {
		errList = append(errList, err)
	} else {
		params.terminateAllOtherSessions = t
	}

	return params, errList

}

//-----------------------------------------------------------------------------

type changePasswordParams struct {
	oldPassword string
	newPassword string
}

func changeUserPassword(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, http.StatusBadRequest, err)
			return
		}

		params, errList := validateChangePasswordParam(r)
		if len(errList) != 0 {
			writeError(ctx, w, http.StatusBadRequest, errList...)
			return
		}

		userAndSession, ok := auth.UserAndSessionFromContext(ctx)
		utils.Assert(ok, "we should find the user in the context tree, but we did not. something is wrong.")

		err = authRepo.ChangePasswordForAllLoginOptions(ctx, int(userAndSession.UserID), params.oldPassword, params.newPassword)
		if err != nil {
			writeError(ctx, w, return400IfAppErrOr500(err), err)
			return
		}

		writeOperationDoneSuccessfullyJson(ctx, w)
	}
}

func validateChangePasswordParam(r *http.Request) (changePasswordParams, []error) {
	oldPassword := r.FormValue("old_password")
	newPassword := r.FormValue("new_password")

	errList := make([]error, 0, 2)

	errList = append(errList, validatePassword(newPassword)...)
	errList = append(errList, validatePassword(oldPassword)...)

	if len(errList) != 0 {
		return changePasswordParams{}, errList
	}

	params := changePasswordParams{
		oldPassword: oldPassword,
		newPassword: newPassword,
	}

	return params, errList
}

//-----------------------------------------------------------------------------

func validatePassword(password string) []error {
	errList := make([]error, 0, 1)

	if len(password) < auth.PasswordRecommendedLength {
		errList = append(errList, apperr.ErrTooShortPassword)
	}

	return errList
}

// -----------------------------------------------------------------------------

type publicUser struct {
	ID           int32       `json:"id"`
	Username     string      `json:"username"`
	ProfileImage pgtype.Text `json:"profile_image"`
	FirstName    string      `json:"first_name"`
	MiddleName   pgtype.Text `json:"middle_name"`
	LastName     pgtype.Text `json:"last_name"`
}

func NewPublicUserFromAuthUser(u auth.User) publicUser {
	return publicUser{
		ID:           u.ID,
		Username:     u.Username,
		ProfileImage: u.ProfileImage,
		FirstName:    u.FirstName,
		MiddleName:   u.MiddleName,
		LastName:     u.LastName,
	}
}

//-----------------------------------------------------------------------------

type forgetPasswordParams struct {
	LoginMethod auth.LoginMethod
	Email       string
	PhoneNumber utils.PhoneNumber
}

func forgetPassword(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, http.StatusBadRequest, err)
			return
		}

		params, errList := validateForgetPasswordParam(r)
		if len(errList) != 0 {
			writeError(ctx, w, http.StatusBadRequest, errList...)
			return
		}

		id, err := authRepo.ForgetPassword(
			ctx,
			auth.PasswordLoginAccessKey{Email: params.Email, Phone: params.PhoneNumber, LoginMethod: params.LoginMethod},
		)
		if err != nil {
			writeError(ctx, w, return400IfAppErrOr500(err), err)
			return
		}

		response := struct {
			Id string `json:"id"`
		}{
			Id: id.String(),
		}

		writeJson(ctx, w, http.StatusOK, response)
	}
}

func validateForgetPasswordParam(r *http.Request) (forgetPasswordParams, []error) {
	errList := make([]error, 0, 2)

	loginMethodFormStr := r.FormValue("login_method")

	emailFormStr := r.FormValue("email")

	phone := utils.PhoneNumber{
		CountryCode: r.FormValue("country_code"),
		Number:      r.FormValue("phone_number"),
	}

	loginMethod, err := new(auth.LoginMethod).FromString(loginMethodFormStr)
	if err != nil {
		errList = append(errList, err)
	}

	loginMethod.FoldOr(
		func() {
			if !emailvalidator.IsValidEmail(emailFormStr) {
				errList = append(errList, apperr.ErrInvalidEmail)
			}
		},
		func() {
			if !phone.IsValid() {
				errList = append(errList, apperr.ErrInvalidPhoneNumber)
			}
		},
		func() {
			// no op
		},
	)

	if len(errList) != 0 {
		return forgetPasswordParams{}, errList
	}

	params := forgetPasswordParams{
		LoginMethod: *loginMethod,
		PhoneNumber: phone,
		Email:       emailFormStr,
	}

	return params, errList
}

//-----------------------------------------------------------------------------

type resetPasswordParams struct {
	Id          uuid.UUID
	Code        string
	newPassword string
}

func resetPassword(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, http.StatusBadRequest, err)
			return
		}

		params, errList := validateResetPasswordParams(r)
		if len(errList) != 0 {
			writeError(ctx, w, http.StatusBadRequest, errList...)
			return
		}

		err = authRepo.ResetPassword(
			ctx,
			params.Id,
			params.Code,
			params.newPassword,
		)
		if err != nil {
			writeError(ctx, w, return400IfAppErrOr500(err), err)
			return
		}

		writeOperationDoneSuccessfullyJson(ctx, w)
	}
}

func validateResetPasswordParams(r *http.Request) (resetPasswordParams, []error) {
	idFormStr := r.FormValue("id")
	code := r.FormValue("code")
	newPassword := r.FormValue("new_password")

	errList := make([]error, 0, 3)

	errList = append(errList, validatePassword(newPassword)...)

	id, err := uuid.Parse(idFormStr)
	if err != nil {
		errList = append(errList, errors.New("invalid id"))
	}

	if len(code) != auth.OtpCodeLength {
		errList = append(errList, errors.New("invalid code"))
	}

	if len(errList) != 0 {
		return resetPasswordParams{}, errList
	}

	params := resetPasswordParams{
		Id:          id,
		Code:        code,
		newPassword: newPassword,
	}

	return params, errList
}
