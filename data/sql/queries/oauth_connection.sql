-- name: OauthConnectionCreate :one
INSERT INTO oauth_connection (
        provider_name,
        scopes
    )
VALUES (
    @provider_name::text,
    @scopes::text[]
)
RETURNING *;
