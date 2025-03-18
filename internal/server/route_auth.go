package server

import (
	"context"
	"net/http"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter/redis_ratelimiter"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/emailvalidator"
	"github.com/redis/go-redis/v9"
)

func authRouter(ctx context.Context, s *Server) http.Handler {
	mux := http.NewServeMux()

	createAcccountRateLimiter := getCreateAcccountRateLimiter(ctx, s.rdb)
	mux.HandleFunc(
		"POST /create-account",
		middleware.MiddlewareChain(
			s.createTempAccount,
			middleware.ACT_app_x_www_form_urlencoded,
			createAcccountRateLimiter,
		),
	)

	loginAcccountRateLimiter := getLoginRateLimiter(ctx, s.rdb)
	mux.HandleFunc(
		"POST /login",
		middleware.MiddlewareChain(
			s.login,
			middleware.ACT_app_x_www_form_urlencoded,
			loginAcccountRateLimiter,
		),
	)

	generalAuthRateLimit := getGeneralAuthRateLimit(ctx, s.rdb)
	return middleware.MiddlewareChain(
		mux.ServeHTTP,
		generalAuthRateLimit,
		s.Installation,
	)
}

type createAccountParams struct {
	LoginMethod auth.LoginMethod
	Email       string
	PhoneNumber utils.PhoneNumber
	Password    string
	FirstName   string
	LastName    string
}

func (s *Server) createTempAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := r.ParseForm()
	if err != nil {
		WriteError(ctx, w, http.StatusBadRequest, err)
		return
	}

	createAccountParam, errList := validateCreateAccountParam(r)
	if len(errList) != 0 {
		WriteError(ctx, w, http.StatusBadRequest, errList...)
		return
	}

	tuser := new(auth.TempUser)
	tuser.LoginMethod = createAccountParam.LoginMethod
	tuser.Email = createAccountParam.Email
	tuser.Phone = createAccountParam.PhoneNumber
	tuser.Password = createAccountParam.Password
	tuser.Lname = createAccountParam.LastName
	tuser.Fname = createAccountParam.FirstName

	authRepo := s.NewAuthRepository()

	tuser, err = authRepo.CreateTempUser(ctx, tuser)
	if err != nil {
		WriteError(ctx, w, http.StatusInternalServerError, err)
		return
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

	errList := make([]error, 5)

	loginMethod, err := new(auth.LoginMethod).FromString(loginMethodFormStr)
	if err != nil {
		errList = append(errList, err)
	}
	if len(passwordFormStr) <= 6 {
		errList = append(errList, apperr.ErrTooShortPassword)
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

type loginParams struct {
	LoginMethod auth.LoginMethod
	Email       string
	PhoneNumber utils.PhoneNumber
	Password    string
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := r.ParseForm()
	if err != nil {
		WriteError(ctx, w, http.StatusBadRequest, err)
		return
	}

	loginParam, errList := validateLoginParam(r)
	if len(errList) != 0 {
		WriteError(ctx, w, http.StatusBadRequest, errList...)
		return
	}

	authRepo := s.NewAuthRepository()
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
		WriteError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	response := struct {
		User  auth.User `json:"user"`
		Token string    `json:"token"`
	}{
		User:  user,
		Token: token,
	}
	WriteJson(ctx, w, http.StatusCreated, response)
}

func validateLoginParam(r *http.Request) (loginParams, []error) {
	loginMethodFormStr := r.FormValue("login_method")

	emailFormStr := r.FormValue("email")

	phone := utils.PhoneNumber{
		CountryCode: r.FormValue("country_code"),
		Number:      r.FormValue("phone_number"),
	}

	passwordFormStr := r.FormValue("password")

	errList := make([]error, 3)

	loginMethod, err := new(auth.LoginMethod).FromString(loginMethodFormStr)
	if err != nil {
		errList = append(errList, err)
	}
	if len(passwordFormStr) <= 6 {
		errList = append(errList, apperr.ErrTooShortPassword)
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

func getCreateAcccountRateLimiter(ctx context.Context, rdb *redis.Client) func(next http.Handler) http.HandlerFunc {
	return middleware.RateLimiter(
		func(r *http.Request) (string, error) {
			return r.RemoteAddr, nil
		},
		redis_ratelimiter.NewRedisSlidingWindowLimiter(
			ctx,
			rdb,
			ratelimiter.Config{
				Disabled:     true,
				PerTimeFrame: 30,
				TimeFrame:    time.Hour * 24,
				KeyPrefix:    "auth:create:account",
			},
		),
	)
}

func getLoginRateLimiter(ctx context.Context, rdb *redis.Client) func(next http.Handler) http.HandlerFunc {
	return middleware.RateLimiter(
		func(r *http.Request) (string, error) {
			err := r.ParseForm()
			if err != nil {
				return "", err
			}
			return r.Form.Get("username"), nil
		},
		redis_ratelimiter.NewRedisSlidingWindowLimiter(
			ctx,
			rdb,
			ratelimiter.Config{
				PerTimeFrame: 10,
				TimeFrame:    time.Hour * 24,
				KeyPrefix:    "auth:login",
			},
		),
	)
}

func getGeneralAuthRateLimit(ctx context.Context, rdb *redis.Client) func(next http.Handler) http.HandlerFunc {
	return middleware.RateLimiter(
		func(r *http.Request) (string, error) {
			return r.RemoteAddr, nil
		},
		redis_ratelimiter.NewRedisSlidingWindowLimiter(
			ctx,
			rdb,
			ratelimiter.Config{
				PerTimeFrame: 60,
				TimeFrame:    time.Hour,
				KeyPrefix:    "auth",
			},
		),
	)
}
