-- name: UsersGetUserById :one
SELECT *
FROM users
WHERE id = $1
    AND deleted_at IS NULL
LIMIT 1;

-- name: UsersGetUserBySessionToken :one
SELECT u.*
FROM session AS s
    JOIN login_option AS lo ON s.originated_from = lo.id
    JOIN users AS u ON u.id = lo.user_id
WHERE s.token = $1
    AND s.deleted_at IS NULL
    AND u.deleted_at IS NULL
    AND lo.deleted_at IS NULL
    AND s.expires_at > NOW()
LIMIT 1;

-- name: UsersIsUsernameUsed :one
SELECT COUNT(*)
FROM users
WHERE username = $1;

-- name: UsersCreateNewUser :one
INSERT INTO users (
        username,
        profile_image,
        first_name,
        last_name,
        role_id
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UsersUpdateUserData :one
UPDATE users
SET username = $2,
    profile_image = $3,
    first_name = $4,
    last_name = $5,
    role_id = $6
WHERE id = $1
RETURNING *;

-- name: UsersUpdateUsernameForUser :exec
UPDATE users
SET username = $2
WHERE id = $1;

-- name: UsersSoftDeleteUser :exec
UPDATE users
SET deleted_at = NOW()
WHERE id = $1;