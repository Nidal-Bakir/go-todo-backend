package auth

import (
	"context"
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

func ContextWithInstallation(ctx context.Context, installation Installation) context.Context {
	return context.WithValue(ctx, currentInstallationCtxKey, installation)
}

func InstallationFromContext(ctx context.Context) (Installation, bool) {
	installation, ok := ctx.Value(currentInstallationCtxKey).(Installation)
	return installation, ok
}
