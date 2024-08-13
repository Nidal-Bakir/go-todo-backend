-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1
LIMIT 1;

-- name: GetUserBySessionToken :one
SELECT u.*
FROM session AS s
    JOIN login_option AS lo ON s.originated_from = lo.id
    JOIN users AS u ON u.id = lo.user_id
WHERE s.token = $1
    AND s.deleted_at IS NULL
    AND s.expires_at > NOW()
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users
VALUES( created_at, updated_at, name)
VALUES ($1, $2, $3, $4)
RETURNING *;