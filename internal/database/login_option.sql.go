// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: login_option.sql

package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createNewLoginOption = `-- name: CreateNewLoginOption :one
INSERT INTO login_option(
        login_method,
        access_key,
        hashed_pass,
        pass_salt,
        verified_at,
        user_id
    )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, login_method, access_key, hashed_pass, pass_salt, verified_at, created_at, updated_at, deleted_at, user_id
`

type CreateNewLoginOptionParams struct {
	LoginMethod string             `json:"login_method"`
	AccessKey   string             `json:"access_key"`
	HashedPass  pgtype.Text        `json:"hashed_pass"`
	PassSalt    pgtype.Text        `json:"pass_salt"`
	VerifiedAt  pgtype.Timestamptz `json:"verified_at"`
	UserID      int32              `json:"user_id"`
}

// CreateNewLoginOption
//
//	INSERT INTO login_option(
//	        login_method,
//	        access_key,
//	        hashed_pass,
//	        pass_salt,
//	        verified_at,
//	        user_id
//	    )
//	VALUES ($1, $2, $3, $4, $5, $6)
//	RETURNING id, login_method, access_key, hashed_pass, pass_salt, verified_at, created_at, updated_at, deleted_at, user_id
func (q *Queries) CreateNewLoginOption(ctx context.Context, arg CreateNewLoginOptionParams) (LoginOption, error) {
	row := q.db.QueryRow(ctx, createNewLoginOption,
		arg.LoginMethod,
		arg.AccessKey,
		arg.HashedPass,
		arg.PassSalt,
		arg.VerifiedAt,
		arg.UserID,
	)
	var i LoginOption
	err := row.Scan(
		&i.ID,
		&i.LoginMethod,
		&i.AccessKey,
		&i.HashedPass,
		&i.PassSalt,
		&i.VerifiedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.UserID,
	)
	return i, err
}

const getActiveLoginOption = `-- name: GetActiveLoginOption :one
SELECT id, login_method, access_key, hashed_pass, pass_salt, verified_at, created_at, updated_at, deleted_at, user_id
FROM login_option
WHERE login_method = $1
    AND access_key = $2
    AND verified_at IS NOT NULL
    AND deleted_at IS NULL
LIMIT 1
`

type GetActiveLoginOptionParams struct {
	LoginMethod string `json:"login_method"`
	AccessKey   string `json:"access_key"`
}

// GetActiveLoginOption
//
//	SELECT id, login_method, access_key, hashed_pass, pass_salt, verified_at, created_at, updated_at, deleted_at, user_id
//	FROM login_option
//	WHERE login_method = $1
//	    AND access_key = $2
//	    AND verified_at IS NOT NULL
//	    AND deleted_at IS NULL
//	LIMIT 1
func (q *Queries) GetActiveLoginOption(ctx context.Context, arg GetActiveLoginOptionParams) (LoginOption, error) {
	row := q.db.QueryRow(ctx, getActiveLoginOption, arg.LoginMethod, arg.AccessKey)
	var i LoginOption
	err := row.Scan(
		&i.ID,
		&i.LoginMethod,
		&i.AccessKey,
		&i.HashedPass,
		&i.PassSalt,
		&i.VerifiedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.UserID,
	)
	return i, err
}

const getActiveLoginOptionWithUser = `-- name: GetActiveLoginOptionWithUser :one
SELECT lo.id as login_option_id,
    lo.login_method as login_option_login_method,
    lo.access_key as login_option_access_key,
    lo.hashed_pass as login_option_hashed_pass,
    lo.pass_salt as login_option_pass_salt,
    lo.verified_at as login_option_verified_at,
    lo.created_at as login_option_created_at,
    lo.updated_at as login_option_updated_at,
    lo.deleted_at as login_option_deleted_at,
    u.id as user_id,
    u.username as user_username,
    u.profile_image as user_profile_image,
    u.first_name as user_first_name,
    u.middle_name as user_middle_name,
    u.last_name as user_last_name,
    u.created_at as user_created_at,
    u.updated_at as user_updated_at,
    u.blocked_at as user_blocked_at,
    u.deleted_at as user_deleted_at,
    u.role_id as user_role_id
FROM login_option AS lo
    JOIN users AS u ON lo.user_id = u.id
WHERE lo.login_method = $1
    AND lo.access_key = $2
    AND lo.verified_at IS NOT NULL
    AND lo.deleted_at IS NULL
    AND u.deleted_at IS NULL
LIMIT 1
`

type GetActiveLoginOptionWithUserParams struct {
	LoginMethod string `json:"login_method"`
	AccessKey   string `json:"access_key"`
}

type GetActiveLoginOptionWithUserRow struct {
	LoginOptionID          int32              `json:"login_option_id"`
	LoginOptionLoginMethod string             `json:"login_option_login_method"`
	LoginOptionAccessKey   string             `json:"login_option_access_key"`
	LoginOptionHashedPass  pgtype.Text        `json:"login_option_hashed_pass"`
	LoginOptionPassSalt    pgtype.Text        `json:"login_option_pass_salt"`
	LoginOptionVerifiedAt  pgtype.Timestamptz `json:"login_option_verified_at"`
	LoginOptionCreatedAt   pgtype.Timestamptz `json:"login_option_created_at"`
	LoginOptionUpdatedAt   pgtype.Timestamptz `json:"login_option_updated_at"`
	LoginOptionDeletedAt   pgtype.Timestamptz `json:"login_option_deleted_at"`
	UserID                 int32              `json:"user_id"`
	UserUsername           string             `json:"user_username"`
	UserProfileImage       pgtype.Text        `json:"user_profile_image"`
	UserFirstName          string             `json:"user_first_name"`
	UserMiddleName         pgtype.Text        `json:"user_middle_name"`
	UserLastName           pgtype.Text        `json:"user_last_name"`
	UserCreatedAt          pgtype.Timestamptz `json:"user_created_at"`
	UserUpdatedAt          pgtype.Timestamptz `json:"user_updated_at"`
	UserBlockedAt          pgtype.Timestamptz `json:"user_blocked_at"`
	UserDeletedAt          pgtype.Timestamptz `json:"user_deleted_at"`
	UserRoleID             pgtype.Int4        `json:"user_role_id"`
}

// GetActiveLoginOptionWithUser
//
//	SELECT lo.id as login_option_id,
//	    lo.login_method as login_option_login_method,
//	    lo.access_key as login_option_access_key,
//	    lo.hashed_pass as login_option_hashed_pass,
//	    lo.pass_salt as login_option_pass_salt,
//	    lo.verified_at as login_option_verified_at,
//	    lo.created_at as login_option_created_at,
//	    lo.updated_at as login_option_updated_at,
//	    lo.deleted_at as login_option_deleted_at,
//	    u.id as user_id,
//	    u.username as user_username,
//	    u.profile_image as user_profile_image,
//	    u.first_name as user_first_name,
//	    u.middle_name as user_middle_name,
//	    u.last_name as user_last_name,
//	    u.created_at as user_created_at,
//	    u.updated_at as user_updated_at,
//	    u.blocked_at as user_blocked_at,
//	    u.deleted_at as user_deleted_at,
//	    u.role_id as user_role_id
//	FROM login_option AS lo
//	    JOIN users AS u ON lo.user_id = u.id
//	WHERE lo.login_method = $1
//	    AND lo.access_key = $2
//	    AND lo.verified_at IS NOT NULL
//	    AND lo.deleted_at IS NULL
//	    AND u.deleted_at IS NULL
//	LIMIT 1
func (q *Queries) GetActiveLoginOptionWithUser(ctx context.Context, arg GetActiveLoginOptionWithUserParams) (GetActiveLoginOptionWithUserRow, error) {
	row := q.db.QueryRow(ctx, getActiveLoginOptionWithUser, arg.LoginMethod, arg.AccessKey)
	var i GetActiveLoginOptionWithUserRow
	err := row.Scan(
		&i.LoginOptionID,
		&i.LoginOptionLoginMethod,
		&i.LoginOptionAccessKey,
		&i.LoginOptionHashedPass,
		&i.LoginOptionPassSalt,
		&i.LoginOptionVerifiedAt,
		&i.LoginOptionCreatedAt,
		&i.LoginOptionUpdatedAt,
		&i.LoginOptionDeletedAt,
		&i.UserID,
		&i.UserUsername,
		&i.UserProfileImage,
		&i.UserFirstName,
		&i.UserMiddleName,
		&i.UserLastName,
		&i.UserCreatedAt,
		&i.UserUpdatedAt,
		&i.UserBlockedAt,
		&i.UserDeletedAt,
		&i.UserRoleID,
	)
	return i, err
}

const markUserLoginOptionAsVerified = `-- name: MarkUserLoginOptionAsVerified :exec
UPDATE login_option
SET verified_at = NOW()
WHERE id = $1
`

// MarkUserLoginOptionAsVerified
//
//	UPDATE login_option
//	SET verified_at = NOW()
//	WHERE id = $1
func (q *Queries) MarkUserLoginOptionAsVerified(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, markUserLoginOptionAsVerified, id)
	return err
}

const setPasswordForUserLoginOption = `-- name: SetPasswordForUserLoginOption :exec
UPDATE login_option
SET hashed_pass = $2,
    pass_salt = $3
WHERE id = $1
`

type SetPasswordForUserLoginOptionParams struct {
	ID         int32       `json:"id"`
	HashedPass pgtype.Text `json:"hashed_pass"`
	PassSalt   pgtype.Text `json:"pass_salt"`
}

// SetPasswordForUserLoginOption
//
//	UPDATE login_option
//	SET hashed_pass = $2,
//	    pass_salt = $3
//	WHERE id = $1
func (q *Queries) SetPasswordForUserLoginOption(ctx context.Context, arg SetPasswordForUserLoginOptionParams) error {
	_, err := q.db.Exec(ctx, setPasswordForUserLoginOption, arg.ID, arg.HashedPass, arg.PassSalt)
	return err
}

const softDeleteUserLoginOption = `-- name: SoftDeleteUserLoginOption :exec
UPDATE login_option
SET deleted_at = NOW()
WHERE id = $1
`

// SoftDeleteUserLoginOption
//
//	UPDATE login_option
//	SET deleted_at = NOW()
//	WHERE id = $1
func (q *Queries) SoftDeleteUserLoginOption(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, softDeleteUserLoginOption, id)
	return err
}
