-- name: OauthTokenUpdate :exec
UPDATE oauth_token
SET access_token = @access_token,
    refresh_token = sqlc.narg(refresh_token)::text,
    token_type = sqlc.narg(token_type)::text,
    expires_at = @expires_at,
    issued_at = @issued_at
WHERE id = @id;
