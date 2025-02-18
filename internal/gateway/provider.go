package gateway

import (
	"context"
)

type Sender interface {
	Send(ctx context.Context, target, content string) error
}

type Provider interface {
	GetSMSProvider(ctx context.Context, contryCode string) Sender
	GetEmailProvider(ctx context.Context) Sender
}

func NewGatewaysProvider(ctx context.Context) Provider {
	return new(providerImpl)
}

type providerImpl struct {
}

func (p providerImpl) GetSMSProvider(ctx context.Context, contryCode string) Sender {
	return newSMSProvider(contryCode)
}

func (p providerImpl) GetEmailProvider(ctx context.Context) Sender {
	return newEmailProvider()
}
