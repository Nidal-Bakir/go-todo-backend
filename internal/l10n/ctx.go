package l10n

import (
	"context"
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
