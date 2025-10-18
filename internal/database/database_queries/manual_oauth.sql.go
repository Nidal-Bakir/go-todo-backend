package database_queries

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const oauthCreateConnectionWithIntegrationDataAndTokens = `-- name: OauthCreateConnectionWithIntegrationDataAndTokens :exec
WITH oauth_connection_record AS (
    SELECT
          $2::text AS provider_name,
          $3::text[] AS scopes
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
        WHERE provider_name = (SELECT provider_name from oauth_connection_record)
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
	SELECT
	    (SELECT id FROM new_oauth_integration),
	    $4::text,
	    $5::text,
	    $6::text,
	    $7::timestamp,
	    $8::timestamp
	WHERE (
	    $4::text IS NOT NULL AND $4::text <> ''
	) OR (
	    $5::text IS NOT NULL AND $5::text <> ''
	)
)
INSERT INTO user_integration (
    oauth_integration_id,
    user_id
)
VALUES (
    (SELECT id FROM new_oauth_integration),
    $1::int
)
`

type OauthCreateConnectionWithIntegrationDataAndTokensParams struct {
	UserID       int32            `json:"user_id"`
	ProviderName string           `json:"provider_name"`
	Scopes       []string         `json:"scopes"`
	AccessToken  pgtype.Text      `json:"access_token"`
	RefreshToken pgtype.Text      `json:"refresh_token"`
	TokenType    pgtype.Text      `json:"token_type"`
	ExpiresAt    pgtype.Timestamp `json:"expires_at"`
	IssuedAt     pgtype.Timestamp `json:"issued_at"`
}

// OauthCreateConnectionWithIntegrationDataAndTokens
//
//	WITH oauth_connection_record AS (
//	    SELECT
//	          $2::text AS provider_name,
//	          $3::text[] AS scopes
//	),
//	oauth_connection_record_merge_op AS (
//	    MERGE INTO oauth_connection AS target
//	    USING oauth_connection_record AS r
//	    ON target.provider_name = r.provider_name AND target.scopes = r.scopes
//	    WHEN NOT MATCHED THEN
//	        INSERT (provider_name, scopes)
//	        VALUES (r.provider_name, r.scopes)
//	    RETURNING target.*
//	),
//	oauth_connection_row AS (
//	    SELECT * FROM oauth_connection_record_merge_op
//	    UNION ALL
//	    SELECT id, provider_name, scopes, created_at, updated_at, deleted_at from oauth_connection
//	        WHERE provider_name = (SELECT provider_name from oauth_connection_record)
//	            AND scopes = (SELECT scopes from oauth_connection_record)
//	),
//	new_oauth_integration AS (
//	    INSERT INTO oauth_integration (
//	        oauth_connection_id,
//	        integration_type
//	    )
//	    VALUES (
//	        (SELECT id FROM oauth_connection_row),
//	        'user'
//	    )
//	    RETURNING id
//	),
//	new_oauth_token AS (
//		INSERT INTO oauth_token (
//		    oauth_integration_id,
//		    access_token,
//		    refresh_token,
//		    token_type,
//		    expires_at,
//		    issued_at
//		)
//		SELECT
//		    (SELECT id FROM new_oauth_integration),
//		    $4::text,
//		    $5::text,
//		    $6::text,
//		    $7::timestamp,
//		    $8::timestamp
//		WHERE (
//		    $4::text IS NOT NULL AND $4::text <> ''
//		) OR (
//		    $5::text IS NOT NULL AND $5::text <> ''
//		)
//	)
//	INSERT INTO user_integration (
//	    oauth_integration_id,
//	    user_id
//	)
//	VALUES (
//	    (SELECT id FROM new_oauth_integration),
//	    $1::int
//	)
func (q *Queries) OauthCreateConnectionWithIntegrationDataAndTokens(ctx context.Context, arg OauthCreateConnectionWithIntegrationDataAndTokensParams) error {
	_, err := q.db.Exec(ctx, oauthCreateConnectionWithIntegrationDataAndTokens,
		arg.UserID,
		arg.ProviderName,
		arg.Scopes,
		arg.AccessToken,
		arg.RefreshToken,
		arg.TokenType,
		arg.ExpiresAt,
		arg.IssuedAt,
	)
	return err
}
