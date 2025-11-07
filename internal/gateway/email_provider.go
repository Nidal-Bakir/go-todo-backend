package gateway

import (
	"context"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/emailvalidator"
	"github.com/rs/zerolog"
)

type simpleEmailProvider struct {
}

func (p simpleEmailProvider) Send(ctx context.Context, target, content string) error {
	err := emailvalidator.IsValidEmailErr(target)
	if err != nil {
		return err
	}

	zerolog.Ctx(ctx).Debug().Str("target", target).Str("content", content).Msg("Sending Email")
	return nil
}

func newEmailProvider() Sender {
	return new(simpleEmailProvider)
}
