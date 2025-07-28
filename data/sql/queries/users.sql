-- name: UsersGetUserById :one
SELECT
    u.id,
    u.username,
    u.profile_image,
    u.first_name,
    u.middle_name,
    u.last_name,
    u.blocked_at,
    u.blocked_until,
    u.created_at,
    u.updated_at,
    u.role_id
FROM not_deleted_users AS u
WHERE id = $1
LIMIT 1;


-- name: UsersGetUserAndSessionDataBySessionToken :one
SELECT s.id as session_id,
    s.token as session_token,
    s.originated_from as session_originated_from,
    s.used_installation as session_used_installation,

    u.id as user_id,
    u.username as user_username,
    u.profile_image as user_profile_image,
    u.first_name as user_first_name,
    u.middle_name as user_middle_name,
    u.last_name as user_last_name,
    u.blocked_at as user_blocked_at,
    u.blocked_until as user_blocked_until,
    u.role_id as user_role_id
FROM active_session AS s
    JOIN active_login_identity AS li ON s.originated_from = li.id
    JOIN not_deleted_users AS u ON u.id = li.user_id
WHERE s.token = $1
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
