package user

import (
	"context"
)

type userCtxKeysType int

const (
	currentUserCtxKey userCtxKeysType = iota
)

func ContextWithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, currentUserCtxKey, user)
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(currentUserCtxKey).(User)
	return user, ok
}
