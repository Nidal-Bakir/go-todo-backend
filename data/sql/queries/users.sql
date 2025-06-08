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

-- name: UsersGetUserAndSessionDataBySessionToken :one
SELECT s.id as session_id,
    s.token as session_token,
    s.created_at as session_created_at,
    s.updated_at as session_updated_at,
    s.expires_at as session_expires_at,
    s.deleted_at as session_deleted_at,
    s.originated_from as session_originated_from,
    s.used_installation as session_used_installation,
    u.id as user_id,
    u.username as user_username,
    u.profile_image as user_profile_image,
    u.first_name as user_first_name,
    u.middle_name as user_middle_name,
    u.last_name as user_last_name,
    u.created_at as user_created_at,
    u.updated_at as user_updated_at,
    u.blocked_at as user_blocked_at,
    u.blocked_until as user_blocked_until,
    u.deleted_at as user_deleted_at,
    u.role_id as user_role_id
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