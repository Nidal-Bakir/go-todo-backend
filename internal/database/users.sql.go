// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: users.sql

package database

import (
	"context"
)

const getUserById = `-- name: GetUserById :one
SELECT id, username, profile_image, first_name, last_name, blocked_at, created_at, updated_at, deleted_at, role_id
FROM users
WHERE id = $1
LIMIT 1
`

// GetUserById
//
//	SELECT id, username, profile_image, first_name, last_name, blocked_at, created_at, updated_at, deleted_at, role_id
//	FROM users
//	WHERE id = $1
//	LIMIT 1
func (q *Queries) GetUserById(ctx context.Context, id int32) (User, error) {
	row := q.db.QueryRow(ctx, getUserById, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.ProfileImage,
		&i.FirstName,
		&i.LastName,
		&i.BlockedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.RoleID,
	)
	return i, err
}

const getUserBySessionToken = `-- name: GetUserBySessionToken :one
SELECT u.id, u.username, u.profile_image, u.first_name, u.last_name, u.blocked_at, u.created_at, u.updated_at, u.deleted_at, u.role_id
FROM session AS s
    JOIN login_option AS lo ON s.originated_from = lo.id
    JOIN users AS u ON u.id = lo.user_id
WHERE s.token = $1
    AND s.deleted_at IS NULL
    AND s.expires_at > NOW()
LIMIT 1
`

// GetUserBySessionToken
//
//	SELECT u.id, u.username, u.profile_image, u.first_name, u.last_name, u.blocked_at, u.created_at, u.updated_at, u.deleted_at, u.role_id
//	FROM session AS s
//	    JOIN login_option AS lo ON s.originated_from = lo.id
//	    JOIN users AS u ON u.id = lo.user_id
//	WHERE s.token = $1
//	    AND s.deleted_at IS NULL
//	    AND s.expires_at > NOW()
//	LIMIT 1
func (q *Queries) GetUserBySessionToken(ctx context.Context, token string) (User, error) {
	row := q.db.QueryRow(ctx, getUserBySessionToken, token)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.ProfileImage,
		&i.FirstName,
		&i.LastName,
		&i.BlockedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.RoleID,
	)
	return i, err
}
