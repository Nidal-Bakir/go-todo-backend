package database_queries

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const loginIdentityCreateNewUserAndOIDCLoginIdentity = `-- name: LoginIdentityCreateNewUserAndOIDCLoginIdentity :one
WITH new_user AS (
  INSERT INTO users (
    username,
    profile_image,
    first_name,
    last_name,
    role_id
  )
  VALUES (
    $1::text,
    $2::text,
    $3::text,
    $4::text,
    $5::int
  )
  RETURNING id AS user_id, username, profile_image, first_name, middle_name, last_name, created_at, updated_at, blocked_at, blocked_until, deleted_at, role_id
),
new_identity AS (
  INSERT INTO login_identity (
    user_id,
    identity_type
  )
  VALUES (
    (SELECT user_id FROM new_user),
    'oidc'
  )
  RETURNING id
),
oauth_provider_record AS (
    SELECT
        $6::text AS provider_name,
        $7::bool AS is_oidc_capable
),
oauth_provider_record_merge_op AS (
    MERGE INTO oauth_provider AS target
    USING oauth_provider_record AS r
    ON target.name = r.provider_name AND target.is_oidc_capable = r.is_oidc_capable
    WHEN NOT MATCHED THEN
        INSERT (name, is_oidc_capable)
        VALUES (r.provider_name, r.is_oidc_capable)
),
oauth_connection_record AS (
    SELECT
        (SELECT provider_name from oauth_provider_record) AS provider_name,
        $8::text[] AS scopes
),
oauth_connection_record_merge_op AS (
    MERGE INTO oauth_connection AS target
    USING oauth_connection_record AS r
    ON target.provider_name = r.provider_name AND target.scopes = r.scopes
    WHEN NOT MATCHED THEN
        INSERT (provider_name, scopes)
        VALUES (r.provider_name, r.scopes)
    RETURNING target.*
),
oauth_connection_row AS (
    SELECT * FROM oauth_connection_record_merge_op
    UNION ALL
    SELECT id, provider_name, scopes, created_at, updated_at, deleted_at from oauth_connection
        WHERE provider_name = (SELECT provider_name from oauth_provider_record)
            AND scopes = (SELECT scopes from oauth_connection_record)
),
new_oauth_integration AS (
    INSERT INTO oauth_integration (
        oauth_connection_id,
        integration_type
    )
    VALUES (
        (SELECT id FROM oauth_connection_row),
        'user'
    )
    RETURNING id
),
new_oauth_token AS (
    INSERT INTO oauth_token (
        oauth_integration_id,
        access_token,
        refresh_token,
        token_type,
        expires_at,
        issued_at
    )
    VALUES (
        (SELECT id FROM new_oauth_integration),
        $9::text,
        $10::text,
        $11::text,
        $12::timestamp,
        $13::timestamp
    )
),
new_user_integration AS (
    INSERT INTO user_integration (
        oauth_integration_id,
        user_id
    )
    VALUES (
        (SELECT id FROM new_oauth_integration),
        (SELECT user_id FROM new_user)
    )
    RETURNING id
),
new_oidc_user_integration_data AS (
    INSERT INTO oidc_user_integration_data (
        user_integration_id,
        sub,
        email,
        iss,
        aud,
        given_name,
        family_name,
        name,
        picture
    )
    VALUES (
        (SELECT id FROM new_user_integration),
        $14::text,
        $15::text,
        $16::text,
        $17::text,
        $18::text,
        $19::text,
        $20::text,
        $21::text
    )
    RETURNING id
),
new_oidc_login_identity AS (
    INSERT INTO oidc_login_identity (
        login_identity_id,
        oidc_user_integration_data_id
    )
    VALUES (
        (SELECT id FROM new_identity),
        (SELECT id FROM new_oidc_user_integration_data)
    )
)
SELECT u.user_id, u.username, u.profile_image, u.first_name, u.middle_name, u.last_name, u.created_at, u.updated_at, u.blocked_at, u.blocked_until, u.deleted_at, u.role_id, i.id AS new_login_identity_id FROM new_user AS u, new_identity AS i
`

type LoginIdentityCreateNewUserAndOIDCLoginIdentityParams struct {
	UserUsername               string           `json:"user_username"`
	UserProfileImage           pgtype.Text      `json:"user_profile_image"`
	UserFirstName              string           `json:"user_first_name"`
	UserLastName               pgtype.Text      `json:"user_last_name"`
	UserRoleID                 pgtype.Int4      `json:"user_role_id"`
	OauthProviderName          string           `json:"oauth_provider_name"`
	OauthProviderIsOidcCapable bool             `json:"oauth_provider_is_oidc_capable"`
	OauthScopes                []string         `json:"oauth_scopes"`
	OauthAccessToken           string           `json:"oauth_access_token"`
	OauthRefreshToken          pgtype.Text      `json:"oauth_refresh_token"`
	OauthTokenType             pgtype.Text      `json:"oauth_token_type"`
	OauthTokenExpiresAt        pgtype.Timestamp `json:"oauth_token_expires_at"`
	OauthTokenIssuedAt         pgtype.Timestamp `json:"oauth_token_issued_at"`
	OidcSub                    string           `json:"oidc_sub"`
	OidcEmail                  pgtype.Text      `json:"oidc_email"`
	OidcIss                    string           `json:"oidc_iss"`
	OidcAud                    string           `json:"oidc_aud"`
	OidcGivenName              pgtype.Text      `json:"oidc_given_name"`
	OidcFamilyName             pgtype.Text      `json:"oidc_family_name"`
	OidcName                   pgtype.Text      `json:"oidc_name"`
	OidcPicture                pgtype.Text      `json:"oidc_picture"`
}

type LoginIdentityCreateNewUserAndOIDCLoginIdentityRow struct {
	UserID             int32              `json:"user_id"`
	Username           string             `json:"username"`
	ProfileImage       pgtype.Text        `json:"profile_image"`
	FirstName          string             `json:"first_name"`
	MiddleName         pgtype.Text        `json:"middle_name"`
	LastName           pgtype.Text        `json:"last_name"`
	CreatedAt          pgtype.Timestamptz `json:"created_at"`
	UpdatedAt          pgtype.Timestamptz `json:"updated_at"`
	BlockedAt          pgtype.Timestamptz `json:"blocked_at"`
	BlockedUntil       pgtype.Timestamptz `json:"blocked_until"`
	DeletedAt          pgtype.Timestamptz `json:"deleted_at"`
	RoleID             pgtype.Int4        `json:"role_id"`
	NewLoginIdentityID int32              `json:"new_login_identity_id"`
}

// LoginIdentityCreateNewUserAndOIDCLoginIdentity
//
// WITH new_user AS (
//   INSERT INTO users (
//     username,
//     profile_image,
//     first_name,
//     last_name,
//     role_id
//   )
//   VALUES (
//     $1::text,
//     $2::text,
//     $3::text,
//     $4::text,
//     $5::int
//   )
//   RETURNING id AS user_id, username, profile_image, first_name, middle_name, last_name, created_at, updated_at, blocked_at, blocked_until, deleted_at, role_id
// ),
// new_identity AS (
//   INSERT INTO login_identity (
//     user_id,
//     identity_type
//   )
//   VALUES (
//     (SELECT user_id FROM new_user),
//     'oidc'
//   )
//   RETURNING id
// ),
// oauth_provider_record AS (
//     SELECT
//         $6::text AS provider_name,
//         $7::bool AS is_oidc_capable
// ),
// oauth_provider_record_merge_op AS (
//     MERGE INTO oauth_provider AS target
//     USING oauth_provider_record AS r
//     ON target.name = r.provider_name AND target.is_oidc_capable = r.is_oidc_capable
//     WHEN NOT MATCHED THEN
//         INSERT (name, is_oidc_capable)
//         VALUES (r.provider_name, r.is_oidc_capable)
// ),
// oauth_connection_record AS (
//     SELECT
//         (SELECT provider_name from oauth_provider_record) AS provider_name,
//         $8::text[] AS scopes
// ),
// oauth_connection_record_merge_op AS (
//     MERGE INTO oauth_connection AS target
//     USING oauth_connection_record AS r
//     ON target.provider_name = r.provider_name AND target.scopes = r.scopes
//     WHEN NOT MATCHED THEN
//         INSERT (provider_name, scopes)
//         VALUES (r.provider_name, r.scopes)
//     RETURNING target.*
// ),
// oauth_connection_row AS (
//     SELECT * FROM oauth_connection_record_merge_op
//     UNION ALL
//     SELECT id, provider_name, scopes, created_at, updated_at, deleted_at from oauth_connection
//         WHERE provider_name = (SELECT provider_name from oauth_provider_record)
//             AND scopes = (SELECT scopes from oauth_connection_record)
// ),
// new_oauth_integration AS (
//     INSERT INTO oauth_integration (
//         oauth_connection_id,
//         integration_type
//     )
//     VALUES (
//         (SELECT id FROM oauth_connection_row),
//         'user'
//     )
//     RETURNING id
// ),
// new_oauth_token AS (
//     INSERT INTO oauth_token (
//         oauth_integration_id,
//         access_token,
//         refresh_token,
//         token_type,
//         expires_at,
//         issued_at
//     )
//     VALUES (
//         (SELECT id FROM new_oauth_integration),
//         $9::text,
//         $10::text,
//         $11::text,
//         $12::timestamp,
//         $13::timestamp
//     )
// ),
// new_user_integration AS (
//     INSERT INTO user_integration (
//         oauth_integration_id,
//         user_id
//     )
//     VALUES (
//         (SELECT id FROM new_oauth_integration),
//         (SELECT user_id FROM new_user)
//     )
//     RETURNING id
// ),
// new_oidc_user_integration_data AS (
//     INSERT INTO oidc_user_integration_data (
//         user_integration_id,
//         sub,
//         email,
//         iss,
//         aud,
//         given_name,
//         family_name,
//         name,
//         picture
//     )
//     VALUES (
//         (SELECT id FROM new_user_integration),
//         $14::text,
//         $15::text,
//         $16::text,
//         $17::text,
//         $18::text,
//         $19::text,
//         $20::text,
//         $21::text
//     )
//     RETURNING id
// ),
// new_oidc_login_identity AS (
//     INSERT INTO oidc_login_identity (
//         login_identity_id,
//         oidc_user_integration_data_id
//     )
//     VALUES (
//         (SELECT id FROM new_identity),
//         (SELECT id FROM new_oidc_user_integration_data)
//     )
// )
// SELECT u.user_id, u.username, u.profile_image, u.first_name, u.middle_name, u.last_name, u.created_at, u.updated_at, u.blocked_at, u.blocked_until, u.deleted_at, u.role_id, i.id AS new_login_identity_id FROM new_user AS u, new_identity AS i
func (q *Queries) LoginIdentityCreateNewUserAndOIDCLoginIdentity(ctx context.Context, arg LoginIdentityCreateNewUserAndOIDCLoginIdentityParams) (LoginIdentityCreateNewUserAndOIDCLoginIdentityRow, error) {
	row := q.db.QueryRow(ctx, loginIdentityCreateNewUserAndOIDCLoginIdentity,
		arg.UserUsername,
		arg.UserProfileImage,
		arg.UserFirstName,
		arg.UserLastName,
		arg.UserRoleID,
		arg.OauthProviderName,
		arg.OauthProviderIsOidcCapable,
		arg.OauthScopes,
		arg.OauthAccessToken,
		arg.OauthRefreshToken,
		arg.OauthTokenType,
		arg.OauthTokenExpiresAt,
		arg.OauthTokenIssuedAt,
		arg.OidcSub,
		arg.OidcEmail,
		arg.OidcIss,
		arg.OidcAud,
		arg.OidcGivenName,
		arg.OidcFamilyName,
		arg.OidcName,
		arg.OidcPicture,
	)
	var i LoginIdentityCreateNewUserAndOIDCLoginIdentityRow
	err := row.Scan(
		&i.UserID,
		&i.Username,
		&i.ProfileImage,
		&i.FirstName,
		&i.MiddleName,
		&i.LastName,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.BlockedAt,
		&i.BlockedUntil,
		&i.DeletedAt,
		&i.RoleID,
		&i.NewLoginIdentityID,
	)
	return i, err
}
