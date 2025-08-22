package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/appenv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth/oauth/google"
	oauth "github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth/oauth/utils"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
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

	mux.HandleFunc("/auth/oidc/{provider}/login", oauthlogin(authRepo))
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

		provider := oauth.ProviderFromString(r.PathValue("provider"))
		if provider == nil {
			http.NotFound(w, r)
			return
		}

		state := oauth.GenerateRandomState()
		verifier := oauth.GenerateVerifier()
		maxAge := int(time.Hour.Seconds())
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			HttpOnly: true,
			Secure:   appenv.IsProd(),
			SameSite: http.SameSiteLaxMode,
			Path:     "/auth/oidc",
			MaxAge:   maxAge,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_verifier",
			Value:    verifier,
			HttpOnly: true,
			Secure:   appenv.IsProd(),
			SameSite: http.SameSiteLaxMode,
			Path:     "/auth/oidc",
			MaxAge:   maxAge,
		})

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

		oidcAndOauthData := auth.LoginOrCreateUserWithOidcRepoParam{
			OauthProvider: *provider,
		}

		err := provider.Fold(oauth.OauthProviderFoldActions{
			OnGoogle: func() error {
				oauthToken, err := google.AuthCodeExchange(r.Context(), code, oauthVerifierFromCookie)
				if err != nil {
					zlog.Err(err).Msg("can not exchange code with tokens")
					return err
				}
				idToken, ok := oauthToken.Extra("id_token").(string)
				if !ok {
					missingIdTokenErr := errors.New("missing id_token from google")
					zlog.Err(missingIdTokenErr).Msg("can not get the id token from the payload")
					return missingIdTokenErr
				}

				oidcAndOauthData.AccessToken = oauthToken.AccessToken
				oidcAndOauthData.RefreshToken = oauthToken.RefreshToken
				oidcAndOauthData.OauthTokenType = oauthToken.TokenType
				oidcAndOauthData.OidcToken = idToken
				oidcAndOauthData.AccessTokenExpiresAt = oauthToken.Expiry
				oidcAndOauthData.OauthScopes = *oauth.NewScops(google.OidcScops)

				return nil
			},
		})
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		installation, ok := auth.InstallationFromContext(ctx)
		utils.Assert(ok, "we should find the installation in the context tree, but we did not. something is wrong.")

		requestIpAddres, err := netip.ParseAddr(r.RemoteAddr)
		if err != nil {
			zlog.Err(err).Msg("can not parse the remoteAddr using netip pkg")
			return
		}

		user, token, err := authRepo.LoginOrCreateUserWithOidc(
			ctx,
			oidcAndOauthData,
			requestIpAddres,
			installation,
		)
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_state",
			Value:    "",
			Path:     "/auth/oidc",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   appenv.IsProd(),
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_verifier",
			Value:    "",
			Path:     "/auth/oidc",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   appenv.IsProd(),
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "Authorization",
			Value:    fmt.Sprint("Bearer ", token),
			HttpOnly: true,
			Secure:   appenv.IsProd(),
			SameSite: http.SameSiteLaxMode,
			Path:     "/auth/oidc",
			MaxAge:   auth.OtpCodeLength,
		})

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
