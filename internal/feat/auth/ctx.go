package auth

import (
	"context"

	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
)

type userCtxKeysType int

const (
	currentUserCtxKey         userCtxKeysType = iota
	currentInstallationCtxKey userCtxKeysType = iota
)

func ContextWithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, currentUserCtxKey, user)
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(currentUserCtxKey).(User)
	return user, ok
}

func ContextWithInstallation(ctx context.Context, installation database.Installation) context.Context {
	return context.WithValue(ctx, currentInstallationCtxKey, installation)
}

func InstallationFromContext(ctx context.Context) (database.Installation, bool) {
	installation, ok := ctx.Value(currentInstallationCtxKey).(database.Installation)
	return installation, ok
}
