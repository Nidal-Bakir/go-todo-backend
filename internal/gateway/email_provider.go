package gateway

import (
	"context"

	"github.com/rs/zerolog"
)

type simpleEmailProvider struct {
}

func (p simpleEmailProvider) Send(ctx context.Context, target, content string) error {
	zlog := zerolog.Ctx(ctx).With().Str("target", target).Str("content", content).Logger()
	zlog.Debug().Msg("Sending Email")
	return nil
}

func newEmailProvider() Sender {
	return new(simpleEmailProvider)
}
