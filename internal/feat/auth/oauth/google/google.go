package google

import (
	"context"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

var (
	clientSecretJsonFileData  []byte
	googleOpenIdConnectConfig *oauth2.Config
	googleIdTokenValidator    *idtoken.Validator
)

func init() {
	clientSecretJsonFileData, err := os.ReadFile("./client_secret_2_557898105577-t939f5gnmonoosmmbjbpid3q6dug21fk.apps.googleusercontent.com.json")
	if err != nil {
		panic(err)
	}
	googleOpenIdConnectConfig, err = google.ConfigFromJSON(clientSecretJsonFileData)
	if err != nil {
		panic(err)
	}
}

func AuthCodeURL(ctx context.Context, state, verifier string) string {
	googleOpenIdConnectConfig.Scopes = []string{"openid", "profile", "email"}

	authUrl := googleOpenIdConnectConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.S256ChallengeOption(verifier),
		oauth2.ApprovalForce,
	)

	return authUrl
}

func AuthCodeExchange(ctx context.Context, code, verifier string) (*oauth2.Token, error) {
	oauth2Token, err := googleOpenIdConnectConfig.Exchange(
		ctx,
		code,
		oauth2.VerifierOption(verifier),
	)
	return oauth2Token, err
}

func ValidatorIdToken(ctx context.Context, idToken string) (*GoogleOidcIdTokenClaims, error) {
	claims := new(GoogleOidcIdTokenClaims)

	if googleIdTokenValidator == nil {
		validator, err := idtoken.NewValidator(
			ctx,
			idtoken.WithCredentialsJSON(clientSecretJsonFileData),
		)
		if err != nil {
			return claims, err
		}
		googleIdTokenValidator = validator
	}

	payload, err := googleIdTokenValidator.Validate(ctx, idToken, googleOpenIdConnectConfig.ClientID)
	if err != nil {
		return claims, err
	}

	claims = newGoogleOidcIdTokenClaimsFromClaims(payload.Claims)

	return claims, nil
}

type GoogleOidcIdTokenClaims struct {
	UserId     string `json:"sub"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	FamilyName string `json:"family_name"`
	GivenName  string `json:"given_name"`
	Picture    string `json:"picture"`
}

func newGoogleOidcIdTokenClaimsFromClaims(claims map[string]any) *GoogleOidcIdTokenClaims {
	c := new(GoogleOidcIdTokenClaims)

	safeConv := func(x any) string {
		if x == nil {
			return ""
		}
		str, ok := x.(string)
		if ok {
			return str
		}
		return ""
	}
	c.UserId = safeConv(claims["sub"])
	c.Email = safeConv(claims["email"])
	c.Name = safeConv(claims["name"])
	c.FamilyName = safeConv(claims["family_name"])
	c.GivenName = safeConv(claims["given_name"])
	c.Picture = safeConv(claims["picture"])

	return c
}
