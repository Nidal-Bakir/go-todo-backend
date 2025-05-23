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

func ContextWithUserAndSession(ctx context.Context, userAndSession UserAndSession) context.Context {
	return context.WithValue(ctx, currentUserCtxKey, userAndSession)
}

func UserAndSessionFromContext(ctx context.Context) (UserAndSession, bool) {
	userAndSession, ok := ctx.Value(currentUserCtxKey).(UserAndSession)
	return userAndSession, ok
}

func ContextWithInstallation(ctx context.Context, installation database.Installation) context.Context {
	return context.WithValue(ctx, currentInstallationCtxKey, installation)
}

func InstallationFromContext(ctx context.Context) (database.Installation, bool) {
	installation, ok := ctx.Value(currentInstallationCtxKey).(database.Installation)
	return installation, ok
}
