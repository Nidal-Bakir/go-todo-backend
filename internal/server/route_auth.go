package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth"
	oauth "github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth/oauth/utils"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter/redis_ratelimiter"
	"github.com/Nidal-Bakir/go-todo-backend/internal/tracker"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/emailvalidator"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/phonenumber"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func authRouter(ctx context.Context, s *Server, authRepo auth.Repository) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"POST /create-account",
		middleware.MiddlewareChain(
			createTempPasswordAccount(authRepo),
			middleware.ACT_app_x_www_form_urlencoded,
			createAcccountRateLimiterByIP(ctx, s.rdb),
			createAcccountRateLimiterByAccessKey(ctx, s.rdb),
			Installation(authRepo),
		),
	)

	mux.HandleFunc(
		"POST /verify-account",
		middleware.MiddlewareChain(
			vareifyAccount(authRepo),
			middleware.ACT_app_x_www_form_urlencoded,
			verifyAccountRateLimiterById(ctx, s.rdb),
			verifyAccountRateLimiterByIP(ctx, s.rdb),
		),
	)

	mux.HandleFunc(
		"POST /login",
		middleware.MiddlewareChain(
			passwordLogin(authRepo),
			middleware.ACT_app_x_www_form_urlencoded,
			loginRateLimiterByAccessKey(ctx, s.rdb),
			loginRateLimiterByIP(ctx, s.rdb),
			Installation(authRepo),
		),
	)

	mux.HandleFunc(
		"POST /logout",
		middleware.MiddlewareChain(
			logout(authRepo),
			middleware.ACT_app_x_www_form_urlencoded,
			Installation(authRepo),
			Auth(authRepo),
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

	mux.HandleFunc(
		"GET /me",
		middleware.MiddlewareChain(
			userProfile(authRepo),
			Auth(authRepo),
		),
	)

	mux.HandleFunc(
		"POST /oidc-login",
		middleware.MiddlewareChain(
			mobileOidcLogin(authRepo),
			middleware.ACT_app_x_www_form_urlencoded,
			mobileOidcLoginRateLimiterByIP(ctx, s.rdb),
			Installation(authRepo),
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
				TimeFrame:    time.Hour * 12,
				KeyPrefix:    "auth:create:account:ip",
			},
		),
	)
}

func createAcccountRateLimiterByAccessKey(ctx context.Context, rdb *redis.Client) func(next http.Handler) http.HandlerFunc {
	return middleware.RateLimiterWithOptionalLimit(
		func(r *http.Request) (key string, shouldRateLimit bool, err error) {
			shouldRateLimit = true

			err = r.ParseForm()
			if err != nil {
				return "", shouldRateLimit, err
			}

			createAccountParam, _ := validateCreateAccountParam(r)
			createAccountParam.LoginIdentityType.FoldOr(
				auth.LoginIdentityFoldActions{
					OnEmail: func() {
						key = createAccountParam.Email
					},
					OnPhone: func() {
						if createAccountParam.PhoneNumber != nil {
							key = createAccountParam.PhoneNumber.ToE164()
						} else {
							// Since the phone number cannot be determined (due to an error or invalid format),
							// we can't apply phone-based rate limiting. Instead, we fall back to IP-based
							// rate limiting and avoid applying the phone-specific limiter for this request.
							shouldRateLimit = false
						}
					},
				},
				func() {
					// Even if the validation has errors, do not return an error.
					// Instead, set the key to the string "unknown_login_identity_type"
					// and allow the misbehaving user to be rate-limited along with other
					// misbehaving users.
					key = "unknown_login_identity_type"
				},
			)
			return key, shouldRateLimit, nil
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

func loginRateLimiterByAccessKey(ctx context.Context, rdb *redis.Client) func(next http.Handler) http.HandlerFunc {
	return middleware.RateLimiterWithOptionalLimit(
		func(r *http.Request) (key string, shouldRateLimit bool, err error) {
			shouldRateLimit = true

			err = r.ParseForm()
			if err != nil {
				return "", shouldRateLimit, err
			}

			loginParam, _ := validatePasswordLoginParam(r)

			loginParam.LoginIdentityType.FoldOr(
				auth.LoginIdentityFoldActions{
					OnEmail: func() {
						key = loginParam.Email
					},
					OnPhone: func() {
						if loginParam.PhoneNumber != nil {
							key = loginParam.PhoneNumber.ToE164()
						} else {
							// Since the phone number cannot be determined (due to an error or invalid format),
							// we can't apply phone-based rate limiting. Instead, we fall back to IP-based
							// rate limiting and avoid applying the phone-specific limiter for this request.
							shouldRateLimit = false
						}
					},
				},
				func() {
					// Even if the validation has errors, do not return an error.
					// Instead, set the key to the string "unknown_login_identity_type"
					// and allow the misbehaving user to be rate-limited along with other
					// misbehaving users.
					key = "unknown_login_identity_type"
				},
			)
			return key, shouldRateLimit, nil
		},
		redis_ratelimiter.NewRedisSlidingWindowLimiter(
			ctx,
			rdb,
			ratelimiter.Config{
				PerTimeFrame: 10,
				TimeFrame:    time.Hour * 12,
				KeyPrefix:    "auth:login:password:identity",
			},
		),
	)
}

func loginRateLimiterByIP(ctx context.Context, rdb *redis.Client) func(next http.Handler) http.HandlerFunc {
	return middleware.RateLimiter(
		func(r *http.Request) (string, error) {
			return r.RemoteAddr, nil
		},
		redis_ratelimiter.NewRedisSlidingWindowLimiter(
			ctx,
			rdb,
			ratelimiter.Config{
				PerTimeFrame: 100,
				TimeFrame:    time.Hour * 12,
				KeyPrefix:    "auth:create:account:ip",
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

func verifyAccountRateLimiterByIP(ctx context.Context, rdb *redis.Client) func(next http.Handler) http.HandlerFunc {
	return middleware.RateLimiter(
		func(r *http.Request) (string, error) {
			return r.RemoteAddr, nil
		},
		redis_ratelimiter.NewRedisSlidingWindowLimiter(
			ctx,
			rdb,
			ratelimiter.Config{
				PerTimeFrame: 25,
				TimeFrame:    time.Hour,
				KeyPrefix:    "auth:verify:account:ip",
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
				TimeFrame:    time.Hour * 12,
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
			param.LoginIdentityType.FoldOr(
				auth.LoginIdentityFoldActions{
					OnEmail: func() {
						key = param.Email
					},
					OnPhone: func() {
						key = param.PhoneNumber.ToE164()
					},
				},
				func() {
					// Even if the validation has errors, do not return an error.
					// Instead, set the key to the string "unknown_login_identity_type"
					// and allow the misbehaving user to be rate-limited along with other
					// misbehaving users.
					key = "unknown_login_identity_type"
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
	LoginIdentityType auth.LoginIdentityType
	Email             string
	PhoneNumber       *phonenumber.PhoneNumber
	Password          string
	FirstName         string
	LastName          string
}

func createTempPasswordAccount(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		createAccountParam, errList := validateCreateAccountParam(r)
		if len(errList) != 0 {
			writeError(ctx, w, r, http.StatusBadRequest, errList...)
			return
		}

		installation := auth.MustInstallationFromContext(ctx)
		if installation.AttachTo != nil {
			err = apperr.ErrInstallationTokenInUse
			writeError(ctx, w, r, return400IfAppErrOr500(err), err)
			return
		}

		tuser := new(auth.TempPasswordUser)
		tuser.LoginIdentityType = createAccountParam.LoginIdentityType
		tuser.Email = createAccountParam.Email
		tuser.Phone = createAccountParam.PhoneNumber
		tuser.Password = createAccountParam.Password
		tuser.Lname = createAccountParam.LastName
		tuser.Fname = createAccountParam.FirstName

		tuser, err = authRepo.CreateTempPasswordUser(ctx, tuser)
		if err != nil {
			writeError(ctx, w, r, return400IfAppErrOr500(err), err)
			return
		}

		response := struct {
			Id string `json:"id"`
		}{
			Id: tuser.Id.String(),
		}

		writeResponse(ctx, w, r, http.StatusCreated, response)
	}
}

func validateCreateAccountParam(r *http.Request) (createAccountParams, []error) {
	createAccountParam := createAccountParams{}
	errList := make([]error, 0, 5)

	loginIdentityTypeFormStr := r.FormValue("login_identity_type")
	loginIdentityType, err := new(auth.LoginIdentityType).FromString(loginIdentityTypeFormStr)
	if err != nil {
		errList = append(errList, err)
		return createAccountParam, errList
	} else {
		createAccountParam.LoginIdentityType = *loginIdentityType
	}

	createAccountParam.Password = r.FormValue("password")
	errList = append(errList, validatePassword(createAccountParam.Password)...)

	loginIdentityType.FoldOr(
		auth.LoginIdentityFoldActions{
			OnEmail: func() {
				createAccountParam.Email = r.FormValue("email")
				if !emailvalidator.IsValidEmail(createAccountParam.Email) {
					errList = append(errList, apperr.ErrInvalidEmail)
				}
			},
			OnPhone: func() {
				phone, err := phonenumber.ParseAndValidate(assumablePhoneNumberFromRequest(r))
				if err != nil {
					errList = append(errList, apperr.ErrInvalidPhoneNumber)
				} else {
					createAccountParam.PhoneNumber = phone
				}
			},
		},
		func() {
			errList = append(errList, apperr.ErrUnsupportedLoginIdentityType)
		},
	)

	createAccountParam.FirstName = r.FormValue("f_name")
	if len(createAccountParam.FirstName) <= 2 {
		errList = append(errList, apperr.ErrTooShortName)
	}
	createAccountParam.LastName = r.FormValue("l_name")
	if len(createAccountParam.LastName) <= 2 {
		errList = append(errList, apperr.ErrTooShortName)
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
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		vareifyAccountParam, errList := validateVareifyAccountParams(r)
		if len(errList) != 0 {
			writeError(ctx, w, r, http.StatusBadRequest, errList...)
			return
		}

		user, err := authRepo.CreatePasswordUser(ctx, vareifyAccountParam.Id, vareifyAccountParam.Code)

		if err != nil {
			writeError(ctx, w, r, return400IfAppErrOr500(err), err)
			return
		}

		response := struct {
			User publicUser `json:"user"`
		}{
			User: NewPublicUserFromAuthUser(user),
		}

		writeResponse(ctx, w, r, http.StatusCreated, response)
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
		errList = append(errList, apperr.ErrInvalidOtpCode)
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

type passwordLoginParams struct {
	LoginIdentityType auth.LoginIdentityType
	Email             string
	PhoneNumber       *phonenumber.PhoneNumber
	Password          string
}

func passwordLogin(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		loginParam, errList := validatePasswordLoginParam(r)
		if len(errList) != 0 {
			writeError(ctx, w, r, http.StatusBadRequest, errList...)
			return
		}

		installation := auth.MustInstallationFromContext(ctx)
		requestIpAddres := tracker.MustReqIPFromContext(ctx)

		user, token, err := authRepo.PasswordLogin(
			ctx,
			auth.PasswordLoginAccessKey{Phone: loginParam.PhoneNumber, Email: loginParam.Email, LoginIdentityType: loginParam.LoginIdentityType},
			loginParam.Password,
			requestIpAddres,
			installation,
		)
		if err != nil {
			statusCode := return400IfAppErrOr500(err)
			if errors.Is(err, apperr.ErrInvalidLoginCredentials) {
				statusCode = http.StatusUnauthorized
			}
			writeError(ctx, w, r, statusCode, err)
			return
		}

		if installation.ClientType.IsWeb() {
			setAuthorizationCookie(w, token)
		}

		response := struct {
			User  publicUser `json:"user"`
			Token string     `json:"token"`
		}{
			User:  NewPublicUserFromAuthUser(user),
			Token: token,
		}
		writeResponse(ctx, w, r, http.StatusCreated, response)
	}
}

func validatePasswordLoginParam(r *http.Request) (passwordLoginParams, []error) {
	loginParam := passwordLoginParams{}
	errList := make([]error, 0, 3)

	loginIdentityTypeFormStr := r.FormValue("login_identity_type")
	loginIdentityType, err := new(auth.LoginIdentityType).FromString(loginIdentityTypeFormStr)
	if err != nil {
		errList = append(errList, err)
		return loginParam, errList
	}
	loginParam.LoginIdentityType = *loginIdentityType

	loginParam.Password = r.FormValue("password")
	errList = append(errList, validatePassword(loginParam.Password)...)

	loginIdentityType.FoldOr(
		auth.LoginIdentityFoldActions{
			OnEmail: func() {
				emailFormStr := r.FormValue("email")
				if !emailvalidator.IsValidEmail(emailFormStr) {
					errList = append(errList, apperr.ErrInvalidEmail)
				} else {
					loginParam.Email = emailFormStr
				}
			},
			OnPhone: func() {
				phone, err := phonenumber.ParseAndValidate(assumablePhoneNumberFromRequest(r))
				if err != nil {
					errList = append(errList, apperr.ErrInvalidPhoneNumber)
				} else {
					loginParam.PhoneNumber = phone
				}
			},
		},
		func() {
			errList = append(errList, apperr.ErrUnsupportedLoginIdentityType)
		},
	)

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
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		logoutParam, errList := validateLogoutParam(r)
		if len(errList) != 0 {
			writeError(ctx, w, r, http.StatusBadRequest, errList...)
			return
		}

		userAndSession := auth.MustUserAndSessionFromContext(ctx)
		installation := auth.MustInstallationFromContext(ctx)

		err = authRepo.Logout(
			ctx,
			int(userAndSession.UserID),
			int(installation.ID),
			int(userAndSession.SessionID),
			logoutParam.terminateAllOtherSessions,
		)
		if err != nil {
			writeError(ctx, w, r, return400IfAppErrOr500(err), err)
			return
		}

		if installation.ClientType.IsWeb() {
			removeAuthorizationCookie(w)
		}

		apiWriteOperationDoneSuccessfullyJson(ctx, w, r)
	}
}

func validateLogoutParam(r *http.Request) (logoutParams, []error) {
	params := logoutParams{}
	errList := make([]error, 0, 1)

	if t, err := strconv.ParseBool(r.FormValue("terminate_all_other_sessions")); err == nil {
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
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		params, errList := validateChangePasswordParam(r)
		if len(errList) != 0 {
			writeError(ctx, w, r, http.StatusBadRequest, errList...)
			return
		}

		userAndSession := auth.MustUserAndSessionFromContext(ctx)

		err = authRepo.ChangePasswordForAllPasswordLoginIdentities(ctx, int(userAndSession.UserID), params.oldPassword, params.newPassword)
		if err != nil {
			writeError(ctx, w, r, return400IfAppErrOr500(err), err)
			return
		}

		apiWriteOperationDoneSuccessfullyJson(ctx, w, r)
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
	LoginIdentityType auth.LoginIdentityType
	Email             string
	PhoneNumber       *phonenumber.PhoneNumber
}

func forgetPassword(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		params, errList := validateForgetPasswordParam(r)
		if len(errList) != 0 {
			writeError(ctx, w, r, http.StatusBadRequest, errList...)
			return
		}

		id, err := authRepo.ForgetPassword(
			ctx,
			auth.PasswordLoginAccessKey{Email: params.Email, Phone: params.PhoneNumber, LoginIdentityType: params.LoginIdentityType},
		)
		if err != nil {
			writeError(ctx, w, r, return400IfAppErrOr500(err), err)
			return
		}

		response := struct {
			Id string `json:"id"`
		}{
			Id: id.String(),
		}

		writeResponse(ctx, w, r, http.StatusOK, response)
	}
}

func validateForgetPasswordParam(r *http.Request) (forgetPasswordParams, []error) {
	params := forgetPasswordParams{}
	errList := make([]error, 0, 2)

	loginIdentityTypeFormStr := r.FormValue("login_identity_type")
	loginIdentityType, err := new(auth.LoginIdentityType).FromString(loginIdentityTypeFormStr)
	if err != nil {
		errList = append(errList, err)
		return params, errList
	} else {
		params.LoginIdentityType = *loginIdentityType
	}

	loginIdentityType.FoldOr(
		auth.LoginIdentityFoldActions{
			OnEmail: func() {
				params.Email = r.FormValue("email")
				if !emailvalidator.IsValidEmail(params.Email) {
					errList = append(errList, apperr.ErrInvalidEmail)
				}
			},
			OnPhone: func() {
				phone, err := phonenumber.ParseAndValidate(assumablePhoneNumberFromRequest(r))
				if err != nil {
					errList = append(errList, apperr.ErrInvalidPhoneNumber)
				} else {
					params.PhoneNumber = phone
				}
			},
		},
		func() {
			errList = append(errList, apperr.ErrUnsupportedLoginIdentityType)
		},
	)
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
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		params, errList := validateResetPasswordParams(r)
		if len(errList) != 0 {
			writeError(ctx, w, r, http.StatusBadRequest, errList...)
			return
		}

		err = authRepo.ResetPassword(
			ctx,
			params.Id,
			params.Code,
			params.newPassword,
		)
		if err != nil {
			writeError(ctx, w, r, return400IfAppErrOr500(err), err)
			return
		}

		apiWriteOperationDoneSuccessfullyJson(ctx, w, r)
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
		errList = append(errList, apperr.ErrInvalidOtpCode)
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

//-----------------------------------------------------------------------------

func userProfile(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userAndSession := auth.MustUserAndSessionFromContext(ctx)

		loginIdentities, err := authRepo.GetAllLoginIdentitiesForUser(ctx, int(userAndSession.UserID))
		if err != nil {
			writeError(ctx, w, r, return400IfAppErrOr500(err), err)
			return
		}

		type PublicAPI struct {
			ID                int32           `json:"id"`
			Email             string          `json:"email,omitzero"`
			Phone             *PhonePublicAPI `json:"phone,omitempty"`
			IsVerified        bool            `json:"is_verified,omitzero"`
			LoginIdentityType string          `json:"login_identity_type"`
			OidcProvider      string          `json:"oidc_provider,omitzero"`
		}

		loginIdentitiesPublicApi := make([]PublicAPI, len(loginIdentities))

		for i, lo := range loginIdentities {

			loginIdentitiesPublicApi[i] = PublicAPI{
				ID:                lo.ID,
				Email:             lo.Email,
				Phone:             NewPhonePublicAPI(lo.Phone),
				IsVerified:        lo.IsVerified,
				LoginIdentityType: lo.LoginIdentityType.String(),
				OidcProvider:      lo.OidcProvider,
			}
		}

		type PublicUser struct {
			ID              int32       `json:"id"`
			Username        string      `json:"username"`
			ProfileImage    pgtype.Text `json:"profile_image"`
			FirstName       string      `json:"first_name"`
			MiddleName      pgtype.Text `json:"middle_name"`
			LastName        pgtype.Text `json:"last_name"`
			LoginIdentities []PublicAPI `json:"login_identities"`
		}

		publicUser := PublicUser{
			ID:              userAndSession.UserID,
			Username:        userAndSession.UserUsername,
			ProfileImage:    userAndSession.UserProfileImage,
			FirstName:       userAndSession.UserFirstName,
			MiddleName:      userAndSession.UserMiddleName,
			LastName:        userAndSession.UserLastName,
			LoginIdentities: loginIdentitiesPublicApi,
		}

		writeResponse(ctx, w, r, http.StatusAccepted, publicUser)
	}
}

//-----------------------------------------------------------------------------

func mobileOidcLoginRateLimiterByIP(ctx context.Context, rdb *redis.Client) func(next http.Handler) http.HandlerFunc {
	return middleware.RateLimiter(
		func(r *http.Request) (string, error) {
			return r.RemoteAddr, nil
		},
		redis_ratelimiter.NewRedisSlidingWindowLimiter(
			ctx,
			rdb,
			ratelimiter.Config{
				PerTimeFrame: 100,
				TimeFrame:    time.Hour * 12,
				KeyPrefix:    "auth:login:oidc:ip",
			},
		),
	)
}

type mobileOidcLoginParams struct {
	provider  *oauth.OauthProvider
	code      string
	oidcToken string
}

func mobileOidcLogin(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		oidcParam, errList := validateMobileOidcLoginParam(r)
		if len(errList) != 0 {
			writeError(ctx, w, r, http.StatusBadRequest, errList...)
			return
		}

		zlog := zerolog.Ctx(ctx).With().Str("oauth_provider", oidcParam.provider.String()).Logger()
		ctx = zlog.WithContext(ctx)

		installation := auth.MustInstallationFromContext(ctx)
		requestIpAddres := tracker.MustReqIPFromContext(ctx)

		user, token, err := authRepo.LoginOrCreateUserWithOidc(
			ctx,
			requestIpAddres,
			installation,
			auth.LoginOrCreateUserWithOidcRepoParam{
				OauthProvider: *oidcParam.provider,
				Code:          oidcParam.code,
				OidcToken:     oidcParam.oidcToken,
			},
		)
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		response := struct {
			User  publicUser `json:"user"`
			Token string     `json:"token"`
		}{
			User:  NewPublicUserFromAuthUser(user),
			Token: token,
		}
		writeResponse(ctx, w, r, http.StatusCreated, response)
	}
}

func validateMobileOidcLoginParam(r *http.Request) (mobileOidcLoginParams, []error) {
	errList := make([]error, 0, 3)

	provider := oauth.ProviderFromString(r.FormValue("provider"))
	if provider == nil {
		errList = append(errList, errors.New("unknown oidc provider"))
	}

	code := r.FormValue("code")
	if len(code) > 2500 {
		errList = append(errList, errors.New("the code is too big"))
	}

	oidcToken := r.FormValue("oidc_token")
	if len(oidcToken) == 0 {
		errList = append(errList, errors.New("the oidc_token is required"))
	}

	params := mobileOidcLoginParams{
		code:      code,
		oidcToken: oidcToken,
		provider:  provider,
	}
	return params, errList
}
