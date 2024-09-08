package tracker

import (
	"context"
	"github.com/google/uuid"
)

type trackerCtxKey int

const (
	reqUUIDCtxKey trackerCtxKey = iota
	lastIPCtxKey
	clientIPCtxKey
)

func ContextWithReqUUID(ctx context.Context, uuidVal uuid.UUID) context.Context {
	return context.WithValue(ctx, reqUUIDCtxKey, uuidVal)
}

func ReqUUIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	uuidVal, ok := ctx.Value(reqUUIDCtxKey).(uuid.UUID)
	return uuidVal, ok
}

func ContextWithLastIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, lastIPCtxKey, ip)
}

func LastIPFromContext(ctx context.Context) (string, bool) {
	ip, ok := ctx.Value(lastIPCtxKey).(string)
	return ip, ok
}

func ContextWithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPCtxKey, ip)
}

func ClientIPFromContext(ctx context.Context) (string, bool) {
	ip, ok := ctx.Value(clientIPCtxKey).(string)
	return ip, ok
}
