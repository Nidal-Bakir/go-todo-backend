package utils

import "fmt"

type PhoneNumber struct {
	CountryCode string
	Number      string
}

func (p PhoneNumber) ToString() string {
	return p.CountryCode + p.Number
}

// 963|123456789
//
// DON't change this form befor you migrate your database to the new form
// change it will cause the old stored phone numbers to not be read correctly
func (p PhoneNumber) ToAppStanderdForm() string {
	return fmt.Sprintf("%s|%s", p.CountryCode, p.Number)
}

func (p PhoneNumber) IsValid() bool {
	// TODO: implement this somehow!!
	return true
}
