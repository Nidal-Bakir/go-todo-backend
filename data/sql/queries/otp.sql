-- name: OTPCreateNewOtp :one
INSERT INTO otp (
        code,
        intent,
        otp_for,
        using_session_id,
        expires_at
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: OTPGetActiveOtpByIdPairedWithSessionId :one
SELECT *
FROM otp
WHERE id = $1
    AND using_session_id = $2
    AND hit_count < 3
    AND expires_at > NOW()
    AND deleted_at IS NULL
LIMIT 1;

-- name: OTPRecordHitForOtp :one
UPDATE otp
SET hit_count = hit_count + 1
WHERE id = $1
RETURNING *;

-- name: OTPSoftDeleteOtp :exec
UPDATE otp
SET deleted_at = NOW()
WHERE id = $1;