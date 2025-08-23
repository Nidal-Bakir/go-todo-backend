-- name: InstallationCreateNewInstallation :exec
INSERT INTO installation (
        installation_token,
        notification_token,
        locale,
        timezone_offset_in_minutes,
        device_manufacturer,
        device_os,
        client_type,
        device_os_version,
        app_version
    )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);


-- name: InstallationUpdateInstallation :exec
UPDATE installation
SET notification_token = $2,
    locale = $3,
    timezone_Offset_in_minutes = $4,
    app_version = $5
WHERE installation_token = $1
    AND deleted_at IS NULL;

-- name: InstallationSoftDeleteInstallation :exec
UPDATE installation
SET deleted_at = NOW()
WHERE id = $1;

-- name: InstallationGetInstallationUsingToken :one
SELECT *
FROM installation
WHERE installation_token = $1
    AND deleted_at IS NULL
LIMIT 1;

-- name: InstallationGetInstallationUsingTokenAndWhereAttachTo :one
SELECT *
FROM installation
WHERE installation_token = $1
    AND attach_to = $2
    AND deleted_at IS NULL
LIMIT 1;

-- name: InstallationAttachSessionToInstallationByToken :exec
UPDATE installation
SET attach_to = $2,
    last_attach_to= NULL
WHERE installation_token = $1
    AND attach_to IS NULL;

-- name: InstallationAttachSessionToInstallationById :execrows
UPDATE installation
SET attach_to = $2,
    last_attach_to= NULL
WHERE id = $1
    AND attach_to IS NULL;

-- name: InstallationDetachSessionFromInstallationByToken :exec
UPDATE installation
SET attach_to = NULL,
        last_attach_to = $2
WHERE installation_token = $1;

-- name: InstallationDetachSessionFromInstallationById :exec
UPDATE installation
SET attach_to = NULL,
    last_attach_to = $2
WHERE id = $1;


-- name: InstallationDetachSessionFromInstallationByUserId :exec
UPDATE installation AS i
SET
    attach_to      = NULL,
    last_attach_to = s.id
FROM active_session AS s
JOIN active_login_identity AS li
    ON s.originated_from = li.id
WHERE
    li.user_id            = $1
    AND i.attach_to       = s.id
    AND i.last_attach_to IS NULL
    AND i.deleted_at     IS NULL;
