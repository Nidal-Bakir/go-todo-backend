// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package database

import (
	"context"
)

type Querier interface {
	//GetUserById
	//
	//  SELECT id, username, email, pass_salt, pass, firstname, lastname FROM users WHERE id = $1 LIMIT 1
	GetUserById(ctx context.Context, id int32) (User, error)
}

var _ Querier = (*Queries)(nil)
