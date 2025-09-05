package oidc

import (
	oauth "github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth/oauth/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type OidcData struct {
	UserFirstName    string
	UserLastName     pgtype.Text
	UserProfileImage pgtype.Text
	UserRoleID       pgtype.Int4

	OidcGivenName  pgtype.Text
	OidcFamilyName pgtype.Text
	OidcName       pgtype.Text
	OidcSub        string
	OidcEmail      pgtype.Text
	OidcPicture    pgtype.Text
	OidcIss        string
	OidcAud        string
	OidcIat        pgtype.Timestamp

	OauthScopes         oauth.Scops
	OauthAccessToken    string
	OauthRefreshToken   pgtype.Text
	OauthTokenType      pgtype.Text
	OauthTokenExpiresAt pgtype.Timestamp
}
