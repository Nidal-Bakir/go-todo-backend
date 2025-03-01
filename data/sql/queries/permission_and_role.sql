-- name: PerRollGetAllPermissions :many
SELECT id,
    name,
    created_at,
    updated_at
FROM permission;

-- name: PerRollGetAllRole :many
SELECT id,
    name,
    created_at,
    updated_at
FROM role;

-- name: PerRollGetRoleWithItsPermissions :many
SELECT r.id as role_id,
    r.name as role_name,
    p.id as permission_id,
    p.name as permission_name
FROM role AS r
    JOIN role_permission AS rp ON r.id = rp.role_id
    JOIN permission AS p ON p.id = rp.permission_id
WHERE r.id = $1;

-- name: PerRollCreateNewPermission :one
INSERT INTO permission(name)
VALUES($1)
RETURNING *;

-- name: PerRollCreateNewRole :one
INSERT INTO role(name)
VALUES($1)
RETURNING *;

-- name: PerRollAddPermissionToRole :exec
INSERT INTO role_permission(role_id, permission_id)
VALUES($1, $2);

-- name: PerRollRemovePermissionFromRole :exec
DELETE FROM role_permission
WHERE role_id = $1
    AND permission_id = $2;

-- name: PerRollSoftDeletePermission :exec
UPDATE permission
SET deleted_at = NOW()
WHERE id = $1;

-- name: PerRollSoftDeleteRole :exec
UPDATE role
SET deleted_at = NOW()
WHERE id = $1;