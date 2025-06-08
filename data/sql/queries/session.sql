-- name: SessionCreateNewSession :one
INSERT INTO session (
        token,
        originated_from,
        used_installation,
        expires_at
    )
VALUES ($1, $2, $3, $4)
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
FROM active_login_option AS lo
WHERE
    s.originated_from = lo.id
    AND lo.user_id    = $1;