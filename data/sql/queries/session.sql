-- name: CreateNewSession :one
INSERT INTO session (
        token,
        originated_from,
        installation_id,
        expires_at
    )
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetActiveSessionById :one
SELECT *
FROM session
WHERE id = $1
    AND expires_at > NOW()
    AND deleted_at IS NULL
LIMIT 1;

-- name: GetActiveSessionByToken :one
SELECT *
FROM session
WHERE token = $1
    AND expires_at > NOW()
    AND deleted_at IS NULL
LIMIT 1;

-- name: SoftDeleteSession :exec
UPDATE session
SET deleted_at = NOW()
WHERE id = $1;