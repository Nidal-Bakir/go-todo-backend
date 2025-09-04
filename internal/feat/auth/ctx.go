package auth

import (
	"context"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
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

func MustUserAndSessionFromContext(ctx context.Context) UserAndSession {
	userAndSession, ok := UserAndSessionFromContext(ctx)
	utils.Assert(ok, "we should find the user in the context tree, but we did not. something is wrong.")
	return userAndSession
}

func ContextWithInstallation(ctx context.Context, installation Installation) context.Context {
	return context.WithValue(ctx, currentInstallationCtxKey, installation)
}

func InstallationFromContext(ctx context.Context) (Installation, bool) {
	installation, ok := ctx.Value(currentInstallationCtxKey).(Installation)
	return installation, ok
}

func MustInstallationFromContext(ctx context.Context) Installation {
	installation, ok := InstallationFromContext(ctx)
	utils.Assert(ok, "we should find the installation in the context tree, but we did not. something is wrong.")
	return installation
}
