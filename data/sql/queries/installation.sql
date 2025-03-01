-- name: InstallationCreateNewInstallation :one
INSERT INTO installation (
        installation_id,
        notification_token,
        locale,
        device_manufacturer,
        device_os,
        device_os_version,
        app_version
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: InstallationUpdateInstallation :exec
UPDATE installation
SET notification_token = $3,
    locale = $4,
    timezone_Offset_in_minutes = $5
WHERE id = $1
    AND installation_id = $2
    AND deleted_at IS NULL;

-- name: InstallationSoftDeleteInstallation :exec
UPDATE installation
SET deleted_at = NOW()
WHERE id = $1;