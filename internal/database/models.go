// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package database

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID        int32
	Username  string
	Email     pgtype.Text
	PassSalt  string
	Pass      string
	Firstname string
	Lastname  string
}