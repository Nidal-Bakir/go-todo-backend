-- name: SessionCreateNewSession :one
INSERT INTO session (
        token,
        originated_from,
        used_installation,
        expires_at,
        ip_address
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: SessionGetActiveSessionById :one
SELECT *
FROM active_session
WHERE id = $1
LIMIT 1;

-- name: SessionGetActiveSessionByToken :one
SELECT *
FROM active_session
WHERE token = $1
LIMIT 1;

-- name: SessionSoftDeleteSession :exec
UPDATE session
SET deleted_at = NOW()
WHERE id = $1;


-- name: SessionSoftDeleteAllActiveSessionsForUser :exec
UPDATE active_session AS s
SET deleted_at = NOW()
FROM active_login_identity AS li
WHERE
    s.originated_from = li.id
    AND li.user_id    = $1;