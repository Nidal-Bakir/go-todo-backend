package appjwt

import (
	"crypto/rsa"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	key            *rsa.PrivateKey
	privateKeyPath = os.Getenv("RSA_PEM_PRIVATE_KEY_PATH")
	appName        = os.Getenv("APP_NAME")
)

func init() {
	privateKeyFileBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		panic(err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyFileBytes)
	if err != nil {
		panic(err)
	}
	key = privateKey
}

func NewAppJWT() *appJWT {
	return &appJWT{key: key, signingMethod: jwt.SigningMethodRS512, issuer: appName}
}

type appJWT struct {
	key           *rsa.PrivateKey
	signingMethod *jwt.SigningMethodRSA
	issuer        string
}

type userClaims struct {
	UserId int `json:"user_id"`
	jwt.RegisteredClaims
}

func (a appJWT) GenWithClaims(tokenExpAt time.Time, userId int, subject string) (token string, err error) {
	timeNow := time.Now()
	claims := userClaims{
		userId,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(tokenExpAt),
			IssuedAt:  jwt.NewNumericDate(timeNow),
			NotBefore: jwt.NewNumericDate(timeNow),
			Issuer:    a.issuer,
			Subject:   subject,
			ID:        uuid.NewString(),
		},
	}

	jwtToken := jwt.NewWithClaims(a.signingMethod, claims)

	sToken, err := jwtToken.SignedString(a.key)
	if err != nil {
		return "", err
	}

	return sToken, nil
}

// be carfull the subject shuold match from the signing phase, use "" to skip it
func (a appJWT) VerifyToken(token, subject string) (*userClaims, error) {
	keyFn := func(t *jwt.Token) (any, error) {
		return &a.key.PublicKey, nil
	}

	parserOptions := []jwt.ParserOption{
		jwt.WithValidMethods([]string{a.signingMethod.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(a.issuer),
		jwt.WithSubject(subject),
		jwt.WithIssuedAt(),
	}

	jwtToken, err := jwt.ParseWithClaims(token, &userClaims{}, keyFn, parserOptions...)
	if err != nil {
		return nil, err
	}

	userClamis, ok := jwtToken.Claims.(*userClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return userClamis, nil

}
