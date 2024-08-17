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

-- name: GetActiveLoginOption :one
SELECT *
FROM login_option
WHERE login_method = $1
    AND access_key = $2
    AND verified_at IS NOT NULL
    AND deleted_at IS NULL
LIMIT 1;

-- name: GetActiveLoginOptionWithUser :one
SELECT *
FROM login_option AS lo
    JOIN users AS u ON lo.user_id = u.id
WHERE lo.login_method = $1
    AND lo.access_key = $2
    AND lo.verified_at IS NOT NULL
    AND lo.deleted_at IS NULL
    AND u.deleted_at IS NULL
LIMIT 1;

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