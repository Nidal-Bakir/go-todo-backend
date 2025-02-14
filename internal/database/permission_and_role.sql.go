// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: permission_and_role.sql

package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addPermissionToRole = `-- name: AddPermissionToRole :exec
INSERT INTO role_permission(role_id, permission_id)
VALUES($1, $2)
`

type AddPermissionToRoleParams struct {
	RoleID       int32 `json:"role_id"`
	PermissionID int32 `json:"permission_id"`
}

// AddPermissionToRole
//
//	INSERT INTO role_permission(role_id, permission_id)
//	VALUES($1, $2)
func (q *Queries) AddPermissionToRole(ctx context.Context, arg AddPermissionToRoleParams) error {
	_, err := q.db.Exec(ctx, addPermissionToRole, arg.RoleID, arg.PermissionID)
	return err
}

const createNewPermission = `-- name: CreateNewPermission :one
INSERT INTO permission(name)
VALUES($1)
RETURNING id, name, created_at, updated_at, deleted_at
`

// CreateNewPermission
//
//	INSERT INTO permission(name)
//	VALUES($1)
//	RETURNING id, name, created_at, updated_at, deleted_at
func (q *Queries) CreateNewPermission(ctx context.Context, name string) (Permission, error) {
	row := q.db.QueryRow(ctx, createNewPermission, name)
	var i Permission
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const createNewRole = `-- name: CreateNewRole :one
INSERT INTO role(name)
VALUES($1)
RETURNING id, name, created_at, updated_at, deleted_at
`

// CreateNewRole
//
//	INSERT INTO role(name)
//	VALUES($1)
//	RETURNING id, name, created_at, updated_at, deleted_at
func (q *Queries) CreateNewRole(ctx context.Context, name string) (Role, error) {
	row := q.db.QueryRow(ctx, createNewRole, name)
	var i Role
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const getAllPermissions = `-- name: GetAllPermissions :many
SELECT id,
    name,
    created_at,
    updated_at
FROM permission
`

type GetAllPermissionsRow struct {
	ID        int32              `json:"id"`
	Name      string             `json:"name"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

// GetAllPermissions
//
//	SELECT id,
//	    name,
//	    created_at,
//	    updated_at
//	FROM permission
func (q *Queries) GetAllPermissions(ctx context.Context) ([]GetAllPermissionsRow, error) {
	rows, err := q.db.Query(ctx, getAllPermissions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetAllPermissionsRow{}
	for rows.Next() {
		var i GetAllPermissionsRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllRole = `-- name: GetAllRole :many
SELECT id,
    name,
    created_at,
    updated_at
FROM role
`

type GetAllRoleRow struct {
	ID        int32              `json:"id"`
	Name      string             `json:"name"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

// GetAllRole
//
//	SELECT id,
//	    name,
//	    created_at,
//	    updated_at
//	FROM role
func (q *Queries) GetAllRole(ctx context.Context) ([]GetAllRoleRow, error) {
	rows, err := q.db.Query(ctx, getAllRole)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetAllRoleRow{}
	for rows.Next() {
		var i GetAllRoleRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRoleWithItsPermissions = `-- name: GetRoleWithItsPermissions :many
SELECT r.id as role_id,
    r.name as role_name,
    p.id as permission_id,
    p.name as permission_name
FROM role AS r
    JOIN role_permission AS rp ON r.id = rp.role_id
    JOIN permission AS p ON p.id = rp.permission_id
WHERE r.id = $1
`

type GetRoleWithItsPermissionsRow struct {
	RoleID         int32  `json:"role_id"`
	RoleName       string `json:"role_name"`
	PermissionID   int32  `json:"permission_id"`
	PermissionName string `json:"permission_name"`
}

// GetRoleWithItsPermissions
//
//	SELECT r.id as role_id,
//	    r.name as role_name,
//	    p.id as permission_id,
//	    p.name as permission_name
//	FROM role AS r
//	    JOIN role_permission AS rp ON r.id = rp.role_id
//	    JOIN permission AS p ON p.id = rp.permission_id
//	WHERE r.id = $1
func (q *Queries) GetRoleWithItsPermissions(ctx context.Context, id int32) ([]GetRoleWithItsPermissionsRow, error) {
	rows, err := q.db.Query(ctx, getRoleWithItsPermissions, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetRoleWithItsPermissionsRow{}
	for rows.Next() {
		var i GetRoleWithItsPermissionsRow
		if err := rows.Scan(
			&i.RoleID,
			&i.RoleName,
			&i.PermissionID,
			&i.PermissionName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const removePermissionFromRole = `-- name: RemovePermissionFromRole :exec
DELETE FROM role_permission
WHERE role_id = $1
    AND permission_id = $2
`

type RemovePermissionFromRoleParams struct {
	RoleID       int32 `json:"role_id"`
	PermissionID int32 `json:"permission_id"`
}

// RemovePermissionFromRole
//
//	DELETE FROM role_permission
//	WHERE role_id = $1
//	    AND permission_id = $2
func (q *Queries) RemovePermissionFromRole(ctx context.Context, arg RemovePermissionFromRoleParams) error {
	_, err := q.db.Exec(ctx, removePermissionFromRole, arg.RoleID, arg.PermissionID)
	return err
}

const softDeletePermission = `-- name: SoftDeletePermission :exec
UPDATE permission
SET deleted_at = NOW()
WHERE id = $1
`

// SoftDeletePermission
//
//	UPDATE permission
//	SET deleted_at = NOW()
//	WHERE id = $1
func (q *Queries) SoftDeletePermission(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, softDeletePermission, id)
	return err
}

const softDeleteRole = `-- name: SoftDeleteRole :exec
UPDATE role
SET deleted_at = NOW()
WHERE id = $1
`

// SoftDeleteRole
//
//	UPDATE role
//	SET deleted_at = NOW()
//	WHERE id = $1
func (q *Queries) SoftDeleteRole(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, softDeleteRole, id)
	return err
}
