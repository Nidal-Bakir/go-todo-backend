package server

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/appenv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth/oauth/google"
	oauth "github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth/oauth/utils"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware"
	"github.com/Nidal-Bakir/go-todo-backend/internal/tracker"
	"github.com/rs/zerolog"
)

func webRouter(_ context.Context, authRepo auth.Repository) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"/",
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./public/index.html")
		},
	)

	mux.Handle(
		"/public/",
		http.StripPrefix(
			"/public/",
			http.FileServer(http.Dir("public")),
		),
	)

	mux.HandleFunc("/auth/oidc/{provider}/login",
		middleware.MiddlewareChain(
			oauthlogin(authRepo),
			Installation(authRepo),
		))
	mux.HandleFunc("/auth/oidc/{provider}/callback",
		middleware.MiddlewareChain(
			oauthloginCallback(authRepo),
			Installation(authRepo),
		),
	)

	return mux
}

func oauthlogin(_ auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		installation := auth.MustInstallationFromContext(ctx)
		if installation.AttachTo != nil {
			writeError(ctx, w, r, http.StatusBadRequest, apperr.ErrInstallationTokenInUse)
			return
		}

		provider := oauth.ProviderFromString(r.PathValue("provider"))
		if provider == nil {
			http.NotFound(w, r)
			return
		}

		state := oauth.GenerateRandomState()
		verifier := oauth.GenerateVerifier()
		setOauthStateAndVerifierCookies(w, state, verifier)

		var redirectUrl string
		provider.Fold(oauth.OauthProviderFoldActions{
			OnGoogle: func() error {
				redirectUrl = google.AuthCodeURL(ctx, state, verifier)
				return nil
			},
		})
		http.Redirect(w, r, redirectUrl, http.StatusFound)
	}
}

func setOauthStateAndVerifierCookies(w http.ResponseWriter, state, verifier string) {
	maxAge := int(time.Hour.Seconds())
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		Secure:   appenv.IsProdOrStag(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/auth/oidc",
		MaxAge:   maxAge,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_verifier",
		Value:    verifier,
		HttpOnly: true,
		Secure:   appenv.IsProdOrStag(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/auth/oidc",
		MaxAge:   maxAge,
	})
}

func removeOauthStateAndVerifierCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/auth/oidc",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   appenv.IsProdOrStag(),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_verifier",
		Value:    "",
		Path:     "/auth/oidc",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   appenv.IsProdOrStag(),
	})
}

func oauthloginCallback(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		provider := oauth.ProviderFromString(r.PathValue("provider"))
		if provider == nil {
			http.NotFound(w, r)
			return
		}

		zlog := zerolog.Ctx(ctx).With().Str("oauth_provider", provider.String()).Logger()
		ctx = zlog.WithContext(ctx)

		oauthStateCookie, oauthStateCookieErr := r.Cookie("oauth_state")
		oauthVerifierCookie, oauthVerifierCookieErr := r.Cookie("oauth_verifier")
		if oauthStateCookieErr != nil || oauthVerifierCookieErr != nil {
			writeError(ctx, w, r, http.StatusBadRequest, errors.New("missing Cookies"))
			zlog.Warn().Msg("missing oauth_state or oauth_verifier cookies")
			return
		}
		oauthStateFromCookie := oauthStateCookie.Value
		oauthVerifierFromCookie := oauthVerifierCookie.Value

		removeOauthStateAndVerifierCookies(w)

		code := r.FormValue("code")
		if len(code) == 0 {
			writeError(ctx, w, r, http.StatusBadRequest, errors.New("missing code"))
			zlog.Warn().Msg("missing the code query parameter from provider redirect request")
			return
		}

		callbackState := r.FormValue("state")
		if callbackState != oauthStateFromCookie {
			writeError(ctx, w, r, http.StatusBadRequest, errors.New("bad state"))
			zlog.Warn().Msg("bad state, the cookie state not equal the callback redirect state from the oauth provider")
			return
		}

		installation := auth.MustInstallationFromContext(ctx)
		requestIpAddres := tracker.MustReqIPFromContext(ctx)
		user, token, err := authRepo.LoginOrCreateUserWithOidc(
			ctx,
			requestIpAddres,
			installation,
			auth.LoginOrCreateUserWithOidcRepoParam{
				OauthProvider: *provider,
				Code:          code,
				CodeVerifier:  oauthVerifierFromCookie,
			},
		)
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		setAuthorizationCookie(w, token)

		queryParams := url.Values{}
		queryParams.Add("user_first_name", user.FirstName)
		queryParams.Add("user_last_name", user.LastName.String)
		queryParams.Add("username", user.Username)
		queryParams.Add("profile_image", user.ProfileImage.String)
		queryParams.Add("id", strconv.Itoa(int(user.ID)))
		queryParams.Add("token", token)
		redirectURL := "/" + "?" + queryParams.Encode()

		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
}
