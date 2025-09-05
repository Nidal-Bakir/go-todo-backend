package google

import (
	"context"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

var (
	clientSecretJsonFileData  []byte
	googleOpenIdConnectConfig *oauth2.Config
	googleIdTokenValidator    *idtoken.Validator
	OidcScops                 = []string{"openid", "profile", "email"}
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
	googleOpenIdConnectConfig.Scopes = OidcScops
	authUrl := googleOpenIdConnectConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.S256ChallengeOption(verifier),
		oauth2.ApprovalForce,
	)
	return authUrl
}

func AuthCodeExchange(ctx context.Context, code, verifier string) (*oauth2.Token, error) {
	authCodeOption := make([]oauth2.AuthCodeOption, 1)
	if len(verifier) != 0 {
		authCodeOption[0] = oauth2.VerifierOption(verifier)
	}
	oauth2Token, err := googleOpenIdConnectConfig.Exchange(
		ctx,
		code,
		authCodeOption...,
	)
	return oauth2Token, err
}

// parse the idTokne without validating it with google
func ParseIdToken(ctx context.Context, idToken string) (*GoogleOidcIdTokenClaims, error) {
	claims := new(GoogleOidcIdTokenClaims)
	payload, err := idtoken.ParsePayload(idToken)
	if err != nil {
		return claims, err
	}
	claims = newGoogleOidcIdTokenClaimsFromIdTokenPayload(payload)
	return claims, nil
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

	claims = newGoogleOidcIdTokenClaimsFromIdTokenPayload(payload)

	return claims, nil
}

type GoogleOidcIdTokenClaims struct {
	Sub        string `json:"sub"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	FamilyName string `json:"family_name"`
	GivenName  string `json:"given_name"`
	Picture    string `json:"picture"`

	Issuer   string    `json:"iss"`
	Audience string    `json:"aud"`
	IssuedAt time.Time `json:"iat"`
}

func newGoogleOidcIdTokenClaimsFromIdTokenPayload(payload *idtoken.Payload) *GoogleOidcIdTokenClaims {
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

	c.Issuer = payload.Issuer
	c.Audience = payload.Audience
	c.IssuedAt = time.Unix(payload.IssuedAt, 0)

	c.Sub = safeConv(payload.Claims["sub"])
	c.Email = safeConv(payload.Claims["email"])
	c.Name = safeConv(payload.Claims["name"])
	c.FamilyName = safeConv(payload.Claims["family_name"])
	c.GivenName = safeConv(payload.Claims["given_name"])
	c.Picture = safeConv(payload.Claims["picture"])

	return c
}
