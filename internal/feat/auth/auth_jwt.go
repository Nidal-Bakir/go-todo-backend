package auth

import (
	"strconv"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/appjwt"
	"github.com/golang-jwt/jwt/v5"
)

const (
	authSubject = "auth"
	userIdKey   = "user_id"

	installationSubject = "installation"
	installationIdKey   = "installation_id"
)

type AuthJWT struct {
	appjwt *appjwt.AppJWT
}

func NewAuthJWT(appjwt *appjwt.AppJWT) *AuthJWT {
	return &AuthJWT{appjwt: appjwt}
}

// ---------------------------------------------------------------------

type AuthClaims struct {
	UserId int32
	jwt.RegisteredClaims
}

func (a AuthClaims) toMap() map[string]string {
	m := make(map[string]string)
	m[userIdKey] = strconv.Itoa(int(a.UserId))
	return m
}

func (authJWT AuthJWT) GenWithClaimsForUser(userId int32, expiresAt time.Time) (string, error) {
	authClaims := AuthClaims{UserId: userId}
	return authJWT.appjwt.GenWithClaims(&expiresAt, authClaims.toMap(), authSubject)
}

func (authJWT AuthJWT) VerifyTokenForUser(token string) (*AuthClaims, error) {
	c, err := authJWT.appjwt.VerifyToken(token, authSubject)
	if err != nil {
		return nil, err
	}

	userId, err := strconv.Atoi(c.Claims[userIdKey])
	if err != nil {
		return nil, err
	}

	return &AuthClaims{UserId: int32(userId), RegisteredClaims: c.RegisteredClaims}, nil
}

// ---------------------------------------------------------------------

type InstallationClaims struct {
	jwt.RegisteredClaims
}

func (a InstallationClaims) toMap() map[string]string {
	m := make(map[string]string)
	return m
}

func (authJWT AuthJWT) GenWithClaimsForInstallation() (string, error) {
	installationClaims := InstallationClaims{}
	return authJWT.appjwt.GenWithClaims(nil, installationClaims.toMap(), installationSubject)
}

func (authJWT AuthJWT) VerifyTokenForInstallation(token string) (*InstallationClaims, error) {
	c, err := authJWT.appjwt.VerifyToken(token, installationSubject)
	if err != nil {
		return nil, err
	}

	return &InstallationClaims{RegisteredClaims: c.RegisteredClaims}, nil
}

// ---------------------------------------------------------------------
