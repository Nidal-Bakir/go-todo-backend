package gateway

import (
	"context"
)

type Sender interface {
	Send(ctx context.Context, target, content string) error
}

type Provider interface {
	NewSMSProvider(ctx context.Context, contryCode string) Sender
	NewEmailProvider(ctx context.Context) Sender
}

func NewGatewaysProvider(ctx context.Context) Provider {
	return new(providerImpl)
}

type providerImpl struct {
}

func (p providerImpl) NewSMSProvider(ctx context.Context, contryCode string) Sender {
	return newSMSProvider(contryCode)
}

func (p providerImpl) NewEmailProvider(ctx context.Context) Sender {
	return newEmailProvider()
}
