package tracker

import (
	"context"

	"github.com/google/uuid"
)

type trackerCtxKey int

const (
	reqUUIDCtxKey trackerCtxKey = iota
)

func ContextWithReqUUID(ctx context.Context, uuidVal uuid.UUID) context.Context {
	return context.WithValue(ctx, reqUUIDCtxKey, uuidVal)
}

func ReqUUIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	uuidVal, ok := ctx.Value(reqUUIDCtxKey).(uuid.UUID)
	return uuidVal, ok
}
