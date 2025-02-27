-- name: CreateNewLoginOption :one
INSERT INTO login_option(
        login_method,
        access_key,
        hashed_pass,
        pass_salt,
        verified_at,
        user_id
    )
VALUES ($1, $2, $3, $4, $5, $6)
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
SELECT lo.id as login_option_id,
    lo.login_method as login_option_login_method,
    lo.access_key as login_option_access_key,
    lo.hashed_pass as login_option_hashed_pass,
    lo.pass_salt as login_option_pass_salt,
    lo.verified_at as login_option_verified_at,
    lo.created_at as login_option_created_at,
    lo.updated_at as login_option_updated_at,
    lo.deleted_at as login_option_deleted_at,
    u.id as user_id,
    u.username as user_username,
    u.profile_image as user_profile_image,
    u.first_name as user_first_name,
    u.middle_name as user_middle_name,
    u.last_name as user_last_name,
    u.created_at as user_created_at,
    u.updated_at as user_updated_at,
    u.blocked_at as user_blocked_at,
    u.deleted_at as user_deleted_at,
    u.role_id as user_role_id
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
