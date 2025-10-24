package oidc

import (
	"context"
	"errors"

	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth/oauth/google"
	oauth "github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth/oauth/utils"
	dbutils "github.com/Nidal-Bakir/go-todo-backend/internal/utils/db_utils"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

func googleOidcFunc(ctx context.Context, code, codeVerifier, oidcToken string) (data OidcData, err error) {
	zlog := zerolog.Ctx(ctx)

	var oauthToken *oauth2.Token
	var claims *google.GoogleOidcIdTokenClaims

	exchangeCode := func() error {
		t, err := google.AuthCodeExchange(ctx, code, codeVerifier)
		if err != nil {
			zlog.Err(err).Msg("can not exchange code with tokens")
			return err
		}

		if scopes := oauth.NewScopesFromOauthToken(t); scopes.Len() == 0 {
			data.OauthScopes = *oauth.NewScopes(google.OidcWebScopes)
		} else {
			data.OauthScopes = *scopes
		}

		oauthToken = t
		return nil
	}

	validateIdTokenFn := func(idToken string) error {
		res, err := google.ValidateIdToken(ctx, idToken)
		if err != nil {
			zlog.Err(err).Msg("can not validate google id token")
			return err
		}
		claims = res
		return nil
	}

	parseIdTokenFn := func(idToken string) error {
		res, err := google.ParseIdToken(ctx, idToken)
		if err != nil {
			zlog.Err(err).Msg("can not parse google id token")
			return err
		}
		claims = res
		return nil
	}

	if len(oidcToken) == 0 {
		err := exchangeCode()
		if err != nil {
			return data, err
		}
		idToken, ok := oauthToken.Extra("id_token").(string)
		if !ok {
			missingIdTokenErr := errors.New("missing id_token from google")
			zlog.Err(missingIdTokenErr).Msg("can not get the id token from the payload")
			return data, missingIdTokenErr
		}
		err = parseIdTokenFn(idToken)
		if err != nil {
			return data, err
		}
	} else {
		err := validateIdTokenFn(oidcToken)
		if err != nil {
			return data, err
		}
		data.OauthScopes = *oauth.NewScopes(google.OidcOpenIdScope)
		if len(code) != 0 {
			err = exchangeCode()
			if err != nil {
				return data, err
			}
		}
	}

	if oauthToken != nil {
		data.OauthAccessToken = dbutils.ToPgTypeText(oauthToken.AccessToken)
		data.OauthRefreshToken = dbutils.ToPgTypeText(oauthToken.RefreshToken)
		data.OauthTokenType = dbutils.ToPgTypeText(oauthToken.TokenType)
		data.OauthTokenExpiresAt = dbutils.ToPgTypeTimestamp(oauthToken.Expiry)
	}

	data.OidcAud = claims.Audience
	data.OidcIss = claims.Issuer
	data.OidcIat = dbutils.ToPgTypeTimestamp(claims.IssuedAt)
	data.OidcSub = claims.Sub
	data.OidcEmail = dbutils.ToPgTypeText(claims.Email)
	data.OidcGivenName = dbutils.ToPgTypeText(claims.GivenName)
	data.OidcFamilyName = dbutils.ToPgTypeText(claims.FamilyName)
	data.OidcName = dbutils.ToPgTypeText(claims.Name)
	data.OidcPicture = dbutils.ToPgTypeText(claims.Picture)
	data.UserFirstName = claims.GivenName
	data.UserLastName = dbutils.ToPgTypeText(claims.FamilyName)
	data.UserProfileImage = dbutils.ToPgTypeText(claims.Picture)

	return data, nil
}
