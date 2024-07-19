// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: users.sql

package database

import (
	"context"
)

const getUserById = `-- name: GetUserById :one
SELECT id, created_at, updated_at, username, email, pass_salt, pass, first_name, last_name, verified_at FROM users WHERE id = $1 LIMIT 1
`

// GetUserById
//
//	SELECT id, created_at, updated_at, username, email, pass_salt, pass, first_name, last_name, verified_at FROM users WHERE id = $1 LIMIT 1
func (q *Queries) GetUserById(ctx context.Context, id int32) (User, error) {
	row := q.db.QueryRow(ctx, getUserById, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Username,
		&i.Email,
		&i.PassSalt,
		&i.Pass,
		&i.FirstName,
		&i.LastName,
		&i.VerifiedAt,
	)
	return i, err
}
