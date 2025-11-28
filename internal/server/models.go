package server

import (
	"fmt"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/phonenumber"
)

type PhonePublicAPI struct {
	CountryCode               string `json:"country_code,omitzero"`
	NationalSignificantNumber string `json:"national_significant_number,omitzero"`
	InternationalFormat       string `json:"international_format,omitzero"`
	E164                      string `json:"e164,omitzero"`
}

func NewPhonePublicAPI(phoneNumber *phonenumber.PhoneNumber) *PhonePublicAPI {
	if phoneNumber == nil {
		return nil
	}
	return &PhonePublicAPI{
		CountryCode:               fmt.Sprint(phoneNumber.CountryCode()),
		NationalSignificantNumber: phoneNumber.NationalSignificantNumber(),
		E164:                      phoneNumber.ToE164(),
		InternationalFormat:       phoneNumber.FormatToInternational(),
	}
}
