package server

import (
	"context"
	"net/http"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter/redis_ratelimiter"
	"github.com/redis/go-redis/v9"
)

func authRouter(ctx context.Context, s *Server) http.Handler {
	mux := http.NewServeMux()

	createAcccountRateLimiter := getCreateAcccountRateLimiter(ctx, s.rdb)
	mux.HandleFunc(
		"POST /create-account",
		middleware.MiddlewareChain(
			s.createAccount,
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

func (s *Server) createAccount(w http.ResponseWriter, r *http.Request) {
	//TODO: implement createAccount
	WriteJson(r.Context(), w, http.StatusCreated, map[string]string{"hi": "nidal"})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := r.ParseForm()
	if err != nil {
		WriteError(ctx, w, http.StatusBadRequest, err)
		return
	}

	// loginParam := struct {
	// 	LoginMethod auth.LoginMethod
	// 	Username    string
	// 	Password    string
	// }{}

	// authRepo := s.NewAuthRepository()
	// installation, ok := auth.InstallationFromContext(ctx)
	// utils.Assert(ok, "we should find the installation in the context tree, but we did not. something is wrong.")

	// user, token, err := authRepo.Login(ctx, loginParam.Username, loginParam.Password, loginParam.LoginMethod, installation)

	// r.Form
	// TODO: not implemented
	//
	//

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
