package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/appenv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth/oauth"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth/oauth/google"
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

	mux.HandleFunc("/auth/oidc/{provider}/login", oauthlogin(authRepo))
	mux.HandleFunc("/auth/oidc/{provider}/callback", oauthloginCallback(authRepo))

	mux.Handle(
		"/public/",
		http.StripPrefix(
			"/public/",
			http.FileServer(http.Dir("public")),
		),
	)

	return mux
}

func oauthlogin(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		provider := r.PathValue("provider")
		if !oauth.IsOauthProviderSupported(provider) {
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

		redirectUrl := google.AuthCodeURL(ctx, state, verifier)
		http.Redirect(w, r, redirectUrl, http.StatusFound)
	}
}

func oauthloginCallback(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		provider := r.PathValue("provider")
		if !oauth.IsOauthProviderSupported(provider) {
			http.NotFound(w, r)
			return
		}

		zlog := zerolog.Ctx(ctx).With().Str("provider", provider).Logger()

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

		oauthToken, err := google.AuthCodeExchange(r.Context(), code, oauthVerifierFromCookie)
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, err)
			zlog.Err(err).Msg("can not exchange code with tokens")
			return
		}

		idToken, ok := oauthToken.Extra("id_token").(string)
		if !ok {
			missingIdTokenErr := errors.New("missing id_token from google")
			writeError(ctx, w, r, http.StatusBadRequest, missingIdTokenErr)
			zlog.Err(missingIdTokenErr).Msg("can not get the id token from the payload")
			return
		}

		userData, err := google.ValidatorIdToken(ctx, idToken)
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, err)
			zlog.Err(err).Msg("can not validate the idToken from google")
			return
		}
		fmt.Println(userData)

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

		http.Redirect(w, r, "/", http.StatusFound)
	}
}
