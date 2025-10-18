-- name: OauthCreateConnectionWithIntegrationDataAndTokens :exec
WITH oauth_connection_record AS (
    SELECT
          @provider_name::text AS provider_name,
          @scopes::text[] AS scopes
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
    SELECT id, provider_name, scopes, created_at, updated_at, deleted_at FROM oauth_connection_record_merge_op
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
		sqlc.narg(access_token)::text,
	    sqlc.narg(refresh_token)::text,
	    sqlc.narg(token_type)::text,
	    @expires_at::timestamp,
	    @issued_at::timestamp
	WHERE (
        sqlc.narg(access_token)::text IS NOT NULL AND sqlc.narg(access_token)::text <> ''
	) OR (
        sqlc.narg(refresh_token)::text IS NOT NULL AND sqlc.narg(refresh_token)::text <> ''
	)
)
INSERT INTO user_integration (
    oauth_integration_id,
    user_id
)
VALUES (
    (SELECT id FROM new_oauth_integration),
    @user_id::int
);
