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
			verifyAccountRateLimiter(ctx, s.rdb),
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
				Disabled:     true,
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
				Disabled:     true,
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

func verifyAccountRateLimiter(ctx context.Context, rdb *redis.Client) func(next http.Handler) http.HandlerFunc {
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
				KeyPrefix:    "auth:verify:account",
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
			writeError(ctx, w, http.StatusInternalServerError, err)
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
			statusCode := http.StatusInternalServerError
			if errors.Is(err, apperr.ErrInvalidOtpCode) || errors.Is(err, apperr.ErrNoResult) {
				statusCode = http.StatusBadRequest
			}
			writeError(ctx, w, statusCode, err)
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
			auth.PasswordLoginAccessKey{Phone: loginParam.PhoneNumber, Email: loginParam.Email},
			loginParam.Password,
			loginParam.LoginMethod,
			installation,
		)
		if err != nil {
			statusCode := http.StatusInternalServerError
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

		err = authRepo.Logout(ctx, int(userAndSession.UserID), int(installation.ID), userAndSession.SessionToken, logoutParam.terminateAllOtherSessions)
		if err != nil {
			writeError(ctx, w, http.StatusBadRequest, err)
			return
		}
	}
}

func validateLogoutParam(r *http.Request) (logoutParams, []error) {
	params := logoutParams{}
	errList := make([]error, 0, 1)

	if t, err := strconv.ParseBool(r.FormValue("terminate_all_other_sessions")); err != nil {
		params.terminateAllOtherSessions = t
	} else {
		errList = append(errList, err)
	}

	return params, errList

}

//-----------------------------------------------------------------------------

type changePasswordParams struct {
	LoginMethod auth.LoginMethod
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
			code := http.StatusInternalServerError
			if _, ok := err.(*apperr.AppErr); ok {
				code = http.StatusBadRequest
			}
			writeError(ctx, w, code, err)
			return
		}

		writeOperationDoneSuccessfullyJson(ctx, w)
	}
}

func validateChangePasswordParam(r *http.Request) (changePasswordParams, []error) {
	oldPassword := r.FormValue("old_password")
	newPassword := r.FormValue("new_password")

	errList := make([]error, 0, 1)

	errList = append(errList, validatePassword(newPassword)...)

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
