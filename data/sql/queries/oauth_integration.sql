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



-- name: OauthIntegrationGetByUserAndScopes :one
SELECT
    ui.id AS user_integration_id,
    ui.user_id AS user_integration_user_id,

    oi.id AS oauth_integration_id,
    oi.id AS oauth_integration_type,

    ot.id AS oauth_token_id,

    oc.id AS oauth_connection_id,
    oc.scopes AS oauth_connection_scopes,
    oc.provider_name AS oauth_connection_provider_name
from active_user_integration AS ui
JOIN active_oauth_integration AS oi
    ON ui.oauth_integration_id = oi.id
JOIN active_oauth_connection AS oc
    ON oi.oauth_connection_id = oc.id
LEFT JOIN active_oauth_token AS ot
    ON ot.oauth_integration_id = oi.id

WHERE ui.user_id = @user_id::INTEGER
    AND oc.scopes = @oauth_scopes::text[]
    AND oc.provider_name = @provider_name::text
LIMIT 1;
