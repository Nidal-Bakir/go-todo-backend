package oauth

import (
	"fmt"
	"slices"

	"golang.org/x/oauth2"
)

type OauthProvider struct {
	idName        string
	isOidcCapable bool
}

var (
	Google   OauthProvider = OauthProvider{idName: "google", isOidcCapable: true}
	Facebook OauthProvider = OauthProvider{idName: "facebook", isOidcCapable: true}
)

func (o OauthProvider) String() string {
	return o.idName
}

var (
	supportedOauthProviders = []OauthProvider{Google, Facebook}
)

func SupportedOauthProviders() []OauthProvider {
	return supportedOauthProviders
}

func ProviderFromString(provider string) *OauthProvider {
	index := slices.IndexFunc(supportedOauthProviders, func(o OauthProvider) bool {
		return o.idName == provider
	})
	if index == -1 {
		return nil
	}
	return &supportedOauthProviders[index]
}

func IsOauthProviderSupported(provider string) bool {
	return slices.ContainsFunc(
		supportedOauthProviders,
		func(p OauthProvider) bool {
			return p.idName == provider
		},
	)
}

func GenerateRandomState() string {
	// will generate random url safe 32 bytes. no need to code a new implementation
	return oauth2.GenerateVerifier()
}

func GenerateVerifier() string {
	return oauth2.GenerateVerifier()
}

type OauthProviderFoldActions struct {
	OnGoogle   func() error
	OnFacebook func() error
}

func (o *OauthProvider) Fold(actions OauthProviderFoldActions) error {
	panicFn := func() error {
		panic(fmt.Sprintf("Not supported oidc provider %s", o.String()))
	}

	if o == nil {
		return panicFn()
	}

	return o.FoldOr(actions, panicFn)
}

func (o *OauthProvider) FoldOr(actions OauthProviderFoldActions, orElse func() error) error {
	if o == nil {
		return orElse()
	}

	actionOrElse := func(fn func() error) func() error {
		if fn == nil {
			return orElse
		}
		return fn
	}

	switch *o {
	case Google:
		return actionOrElse(actions.OnGoogle)()

	case Facebook:
		return actionOrElse(actions.OnFacebook)()

	default:
		return orElse()
	}
}
