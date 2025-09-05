package oidc

import (
	"context"

	oauth "github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth/oauth/utils"
)

type OidcFunc func(ctx context.Context, code, codeVerifier, oidcToken string) (OidcData, error)

func (f OidcFunc) Exec(ctx context.Context, code, codeVerifier, oidcToken string) (OidcData, error) {
	return f(ctx, code, codeVerifier, oidcToken)
}

func NewOidc(provider oauth.OauthProvider) OidcFunc {
	var fn OidcFunc
	provider.Fold(
		oauth.OauthProviderFoldActions{
			OnGoogle: func() error {
				fn = googleOidcFunc
				return nil
			},
		},
	)
	return fn
}
