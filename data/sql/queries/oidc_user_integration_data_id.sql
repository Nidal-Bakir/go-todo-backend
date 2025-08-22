-- name: OidcUserIntegrationDataUpdate :exec
UPDATE oidc_user_integration_data
SET email = sqlc.narg(email)::text,
    given_name = sqlc.narg(given_name)::text,
    family_name = sqlc.narg(family_name)::text,
    name = sqlc.narg(name)::text,
    picture = sqlc.narg(picture)::text
WHERE id = @id;
