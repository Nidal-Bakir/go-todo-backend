-- name: OauthIntegrationUpdateToOauthConnectionBasedOnNewScopes :exec
WITH oauth_connection_record AS (
    SELECT
        @provider_name::text AS provider_name,
        @oauth_scopes::text[] AS scopes
),
oauth_connection_record_merge_op AS (
    MERGE INTO oauth_connection AS target
    USING oauth_connection_record AS r
    ON target.provider_name = r.provider_name AND target.scopes = r.scopes
    WHEN NOT MATCHED THEN
        INSERT (provider_name, scopes)
        VALUES (r.provider_name, r.scopes)
),
oauth_connection_row AS (
    SELECT * from oauth_connection
    WHERE provider_name = @provider_name::text
        AND scopes = @oauth_scopes::text[]
)
UPDATE oauth_integration
SET oauth_connection_id = (SELECT id FROM oauth_connection_row)
WHERE id = @integration_id::int;