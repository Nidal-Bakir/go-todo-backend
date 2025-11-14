-- name: PermGetAllPermissions :many
SELECT
    name,
    created_at,
    updated_at
FROM permission;

-- name: PermGetAllRoles :many
SELECT
    name,
    created_at,
    updated_at
FROM role;

-- name: PermGetRoleWithItsPermissions :many
SELECT
    r.name as role_name,
    p.name as permission_name
FROM role AS r
    JOIN role_permission AS rp ON r.name = rp.role_name
    JOIN permission AS p ON p.name = rp.permission_name
WHERE r.name = $1;

-- name: PermCreateNewPermission :one
INSERT INTO permission(name)
VALUES($1)
RETURNING *;

-- name: PermCreateNewPermissions :copyfrom
INSERT INTO permission(name)
VALUES($1);

-- name: PermCreateNewRole :one
INSERT INTO role(name)
VALUES($1)
RETURNING *;

-- name: PermCreateNewRoles :copyfrom
INSERT INTO role(name)
VALUES($1);

-- name: PermAddPermissionToRole :exec
INSERT INTO role_permission(role_name, permission_name)
VALUES($1, $2);

-- name: PermAddPermissionsToRoles :copyfrom
INSERT INTO role_permission(role_name, permission_name)
VALUES($1, $2);

-- name: PermRemovePermissionFromRole :exec
DELETE FROM role_permission
WHERE role_name = $1
    AND permission_name = $2;

-- name: PermSoftDeletePermission :exec
UPDATE permission
SET deleted_at = NOW()
WHERE name = $1;

-- name: PermSoftDeleteRole :exec
UPDATE role
SET deleted_at = NOW()
WHERE name = $1;
