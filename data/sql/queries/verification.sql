-- name: SoftDeleteVerification :exec
UPDATE verification
SET deleted_at = NOW()
WHERE id = $1;
