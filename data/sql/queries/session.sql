-- name: SessionCreateNewSession :exec
INSERT INTO session (
        token,
        originated_from,
        used_installation,
        expires_at
    )
VALUES ($1, $2, $3, $4);

-- name: SessionGetActiveSessionById :one
SELECT *
FROM session
WHERE id = $1
    AND expires_at > NOW()
    AND deleted_at IS NULL
LIMIT 1;

-- name: SessionGetActiveSessionByToken :one
SELECT *
FROM session
WHERE token = $1
    AND expires_at > NOW()
    AND deleted_at IS NULL
LIMIT 1;

-- name: SessionSoftDeleteSession :exec
UPDATE session
SET deleted_at = NOW()
WHERE id = $1;