-- name: CreateNewLoginOption :one
INSERT INTO login_option(
        login_method,
        access_key,
        hashed_pass,
        pass_salt,
        user_id
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: MarkUserLoginOptionAsVerified :exec
UPDATE login_option
SET verified_at = NOW()
WHERE id = $1;

-- name: SetPasswordForUserLoginOption :exec
UPDATE login_option
SET hashed_pass = $2,
    pass_salt = $3
WHERE id = $1;

-- name: SoftDeleteUserLoginOption :exec
UPDATE login_option
SET deleted_at = NOW()
WHERE id = $1;