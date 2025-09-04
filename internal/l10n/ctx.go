package l10n

import (
	"context"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
)

type l10nCtxKeysType int

const (
	localizerCtxKey l10nCtxKeysType = iota
)

func ContextWithLocalizer(ctx context.Context, localizer *Localizer) context.Context {
	return context.WithValue(ctx, localizerCtxKey, localizer)
}

func LocalizerFromContext(ctx context.Context) (*Localizer, bool) {
	localizer, ok := ctx.Value(localizerCtxKey).(*Localizer)
	return localizer, ok
}

func MustLocalizerFromContext(ctx context.Context) *Localizer {
	localizer, ok := LocalizerFromContext(ctx)
	utils.Assert(ok, "we should find the localizer in the context tree, but we did not. something is wrong.")
	return localizer
}
