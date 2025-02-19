package otp

import (
	"context"
	"math/rand/v2"
	"strings"

	"github.com/Nidal-Bakir/go-todo-backend/internal/gateway"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
)

const (
	otpChars = "0123456789"
)

type OTPSender struct {
	provider gateway.Provider
}

func NewOTPSender(provider gateway.Provider) *OTPSender {
	return &OTPSender{provider: provider}
}

func (o OTPSender) SendSMSOTP(ctx context.Context, target utils.PhoneNumber) (otp string, err error) {
	otp = o.genRandOTP(6)
	err = o.provider.GetSMSProvider(ctx, target.CounterCode).Send(ctx, target.ToString(), otp)
	return otp, err
}

func (o OTPSender) SendEmailOTP(ctx context.Context, target string) (otp string, err error) {
	otp = o.genRandOTP(6)
	err = o.provider.GetEmailProvider(ctx).Send(ctx, target, otp)
	return otp, err
}

func (_ OTPSender) genRandOTP(otpLength uint8) string {
	strBuild := strings.Builder{}
	for range otpLength {
		otpChar := otpChars[rand.IntN(len(otpChars))]
		strBuild.WriteRune(rune(otpChar))
	}
	return strBuild.String()
}
