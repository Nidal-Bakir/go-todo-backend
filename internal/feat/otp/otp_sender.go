package otp

import (
	"context"
	"math/rand/v2"
	"strings"

	"github.com/Nidal-Bakir/go-todo-backend/internal/gateway"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/phonenumber"
)

const (
	otpChars = "0123456789"
)

type OTPSender struct {
	provider  gateway.Provider
	otpLength uint8
}

func NewOTPSender(_ context.Context, provider gateway.Provider, otpLength uint8) *OTPSender {
	utils.Assert(otpLength >= 3, "you can not have otp with length less then 2. What are you doing?")
	return &OTPSender{provider: provider, otpLength: otpLength}
}

func (o OTPSender) SendSmsOtpForAccountVerification(ctx context.Context, target *phonenumber.PhoneNumber) (otp string, err error) {
	otp = o.genRandOTP()
	err = o.sendSmsOtp(ctx, target, otp)
	return otp, err
}

func (o OTPSender) SendEmailOtpForAccountVerification(ctx context.Context, target string) (otp string, err error) {
	otp = o.genRandOTP()
	err = o.sendEmailOtp(ctx, target, otp)
	return otp, err
}

func (o OTPSender) SendSmsOtpForForgetPassword(ctx context.Context, target *phonenumber.PhoneNumber) (otp string, err error) {
	otp = o.genRandOTP()
	err = o.sendSmsOtp(ctx, target, otp)
	return otp, err
}

func (o OTPSender) SendEmailOtpForForgetPassword(ctx context.Context, target string) (otp string, err error) {
	otp = o.genRandOTP()
	err = o.sendEmailOtp(ctx, target, otp)
	return otp, err
}

func (o OTPSender) sendSmsOtp(ctx context.Context, target *phonenumber.PhoneNumber, content string) (err error) {
	return o.provider.NewSMSProvider(ctx, target.CountryCode()).Send(ctx, target.ToE164(), content)
}

func (o OTPSender) sendEmailOtp(ctx context.Context, target string, content string) (err error) {
	return o.provider.NewEmailProvider(ctx).Send(ctx, target, content)
}

func (o OTPSender) genRandOTP() string {
	strBuild := strings.Builder{}
	for range o.otpLength {
		otpChar := otpChars[rand.IntN(len(otpChars))]
		strBuild.WriteRune(rune(otpChar))
	}
	return strBuild.String()
}
