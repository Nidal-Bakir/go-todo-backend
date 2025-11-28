package phonenumber

import (
	"fmt"

	"github.com/nyaruka/phonenumbers"
)

type PhoneNumber struct {
	internal *phonenumbers.PhoneNumber
}

// Should be +xxxxxxxxx (OR E164)
//
// +<CountryCode><NationalNumber>
//
// ex: +963123456789
func Parse(phoneNumberStr string) (*PhoneNumber, error) {
	num, err := phonenumbers.Parse(phoneNumberStr, phonenumbers.UNKNOWN_REGION)
	if err != nil {
		return nil, err
	}
	return &PhoneNumber{internal: num}, nil
}

func MustParse(phoneNumberStr string) *PhoneNumber {
	num, err := Parse(phoneNumberStr)
	if err != nil {
		panic(err)
	}
	return num
}

func MayParse(phoneNumberStr string) *PhoneNumber {
	num, err := Parse(phoneNumberStr)
	if err != nil {
		return nil
	}
	return num
}

func ParseAndValidate(phoneNumberStr string) (*PhoneNumber, error) {
	num, err := Parse(phoneNumberStr)
	if err != nil {
		return nil, err
	}
	if isValid := num.IsValidPhoneNumber(); !isValid {
		return nil, fmt.Errorf("invalid phone number")
	}
	return num, nil
}

func (p *PhoneNumber) IsValidPhoneNumber() bool {
	if p == nil {
		return false
	}
	if !phonenumbers.IsPossibleNumber(p.internal) {
		return false
	}
	return phonenumbers.IsValidNumber(p.internal)
}

func (p *PhoneNumber) ToE164() string {
	return phonenumbers.Format(p.internal, phonenumbers.E164)
}

func (p *PhoneNumber) CountryCode() int {
	return int(p.internal.GetCountryCode())
}

func (p *PhoneNumber) NationalNumber() int {
	return int(p.internal.GetNationalNumber())
}

func (p *PhoneNumber) NationalSignificantNumber() string {
	return phonenumbers.GetNationalSignificantNumber(p.internal)
}

func (p *PhoneNumber) FormatToInternational() string {
	return phonenumbers.Format(p.internal, phonenumbers.INTERNATIONAL)
}
