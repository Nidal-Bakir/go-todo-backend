package gateway

import (
	"context"

	"github.com/rs/zerolog"
)

type simpleSMSProvider struct {
}

func (p simpleSMSProvider) Send(ctx context.Context, target, content string) error {
	zlog := zerolog.Ctx(ctx).With().Str("target", target).Str("content", content).Logger()
	zlog.Debug().Msg("Sending SMS")
	return nil
}

func newSMSProvider(_ string) Sender {
	return new(simpleSMSProvider)
}
