package gateway

import (
	"context"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/rs/zerolog"
)

type simpleEmailProvider struct {
}

func (p simpleEmailProvider) Send(ctx context.Context, target, content string) error {
	err := utils.IsValidEmailErr(target)
	if err != nil {
		return err
	}

	zlog := zerolog.Ctx(ctx).With().Str("target", target).Str("content", content).Logger()
	zlog.Debug().Msg("Sending Email")
	return nil
}

func newEmailProvider() Sender {
	return new(simpleEmailProvider)
}
