package gateway

import (
	"context"

	"github.com/rs/zerolog"
)

type simpleSMSProvider struct {
}

func (p simpleSMSProvider) Send(ctx context.Context, target, content string) error {
	zerolog.Ctx(ctx).Debug().Str("target", target).Str("content", content).Msg("Sending SMS")
	return nil
}

func newSMSProvider(_ int) Sender {
	return new(simpleSMSProvider)
}
