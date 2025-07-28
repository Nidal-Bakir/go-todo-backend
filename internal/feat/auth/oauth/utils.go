package oauth

import (
	"slices"

	"golang.org/x/oauth2"
)

type OauthProvider string

const (
	Google   OauthProvider = "google"
	Facebook OauthProvider = "facebook"
)

var (
	supportedOauthProviders = []OauthProvider{Google, Facebook}
)

func SupportedOauthProviders() []OauthProvider {
	return slices.Clone(supportedOauthProviders)
}

func IsOauthProviderSupported(provider string) bool {
	return slices.Contains(supportedOauthProviders, OauthProvider(provider))
}

func GenerateRandomState() string {
	// will generate random url safe 32 bytes. no need to code a new implementation
	return oauth2.GenerateVerifier()
}

func GenerateVerifier() string {
	return oauth2.GenerateVerifier()
}
