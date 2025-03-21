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
	provider  gateway.Provider
	otpLength uint8
}

func NewOTPSender(provider gateway.Provider, otpLength uint8) *OTPSender {
	utils.Assert(otpLength >= 3, "you can not have otp with length less then 2. What are you doing?")
	return &OTPSender{provider: provider, otpLength: otpLength}
}

func (o OTPSender) SendSMSOTP(ctx context.Context, target utils.PhoneNumber) (otp string, err error) {
	otp = o.genRandOTP()
	err = o.provider.GetSMSProvider(ctx, target.CountryCode).Send(ctx, target.ToString(), otp)
	return otp, err
}

func (o OTPSender) SendEmailOTP(ctx context.Context, target string) (otp string, err error) {
	otp = o.genRandOTP()
	err = o.provider.GetEmailProvider(ctx).Send(ctx, target, otp)
	return otp, err
}

func (o OTPSender) genRandOTP() string {
	strBuild := strings.Builder{}
	for range o.otpLength {
		otpChar := otpChars[rand.IntN(len(otpChars))]
		strBuild.WriteRune(rune(otpChar))
	}
	return strBuild.String()
}
