package tracker

import (
	"context"
	"net/netip"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/google/uuid"
)

type trackerCtxKey int

const (
	ReqIdStrKey   string        = "request_id"
	reqUUIDCtxKey trackerCtxKey = iota
	reqIPCtxKey   trackerCtxKey = iota
)

func ContextWithReqUUID(ctx context.Context, uuidVal uuid.UUID) context.Context {
	return context.WithValue(ctx, reqUUIDCtxKey, uuidVal)
}

func ReqUUIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	uuidVal, ok := ctx.Value(reqUUIDCtxKey).(uuid.UUID)
	return uuidVal, ok
}

func MustReqUUIDFromContext(ctx context.Context) uuid.UUID {
	uuidVal, ok := ReqUUIDFromContext(ctx)
	utils.Assert(ok, "we should find the request uuid in the context tree, but we did not. something is wrong.")
	return uuidVal
}

func ContextWithReqIP(ctx context.Context, ip netip.Addr) context.Context {
	return context.WithValue(ctx, reqIPCtxKey, ip)
}

func ReqIPFromContext(ctx context.Context) (netip.Addr, bool) {
	ip, ok := ctx.Value(reqIPCtxKey).(netip.Addr)
	return ip, ok
}

func MustReqIPFromContext(ctx context.Context) netip.Addr {
	ip, ok := ReqIPFromContext(ctx)
	utils.Assert(ok, "we should find the request IP addres in the context tree, but we did not. something is wrong.")
	return ip
}
