-- name: OauthTokenUpdate :exec
UPDATE oauth_token
SET access_token = sqlc.narg(access_token)::text,
    refresh_token = sqlc.narg(refresh_token)::text,
    token_type = sqlc.narg(token_type)::text,
    expires_at = @expires_at,
    issued_at = @issued_at
WHERE id = @id;


-- name: OauthTokenCreate :exec
INSERT INTO oauth_token (
	    oauth_integration_id,
	    access_token,
	    refresh_token,
	    token_type,
	    expires_at,
	    issued_at
	)
	VALUES (
	    @oauth_integration_id,
	    sqlc.narg(access_token)::text,
        sqlc.narg(refresh_token)::text,
        sqlc.narg(token_type)::text,
        @expires_at,
        @issued_at
	);
