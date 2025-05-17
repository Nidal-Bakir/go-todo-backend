package appjwt

import (
	"crypto/rsa"
	"errors"
	"os"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
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

func NewAppJWT() *AppJWT {
	return &AppJWT{key: key, signingMethod: jwt.SigningMethodRS512, issuer: appName}
}

type AppJWT struct {
	key           *rsa.PrivateKey
	signingMethod *jwt.SigningMethodRSA
	issuer        string
}

type CustomClaims struct {
	Claims map[string]string `json:"claims"`
	jwt.RegisteredClaims
}

func (a AppJWT) GenWithClaims(tokenExpAt time.Time, claims map[string]string, subject string) (token string, err error) {
	timeNow := time.Now()

	customClaims := CustomClaims{
		claims,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(tokenExpAt),
			IssuedAt:  jwt.NewNumericDate(timeNow),
			NotBefore: jwt.NewNumericDate(timeNow),
			Issuer:    a.issuer,
			Subject:   subject,
			ID:        uuid.NewString(),
		},
	}

	jwtToken := jwt.NewWithClaims(a.signingMethod, customClaims)

	sToken, err := jwtToken.SignedString(a.key)
	if err != nil {
		return "", err
	}

	return sToken, nil
}

// be carfull the subject shuold match from the signing phase, use "" to skip it
func (a AppJWT) VerifyToken(token, subject string) (*CustomClaims, error) {
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

	jwtToken, err := jwt.ParseWithClaims(token, &CustomClaims{}, keyFn, parserOptions...)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperr.ErrExpiredSessionToken
		}
		return nil, err
	}

	userClamis, ok := jwtToken.Claims.(*CustomClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return userClamis, nil
}
