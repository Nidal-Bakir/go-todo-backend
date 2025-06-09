package phone_number

import (
	"fmt"
	"strings"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
)

const (
	sep = "|"
)

type PhoneNumber struct {
	CountryCode string
	Number      string
}

func NewPhoneNumberFromStdForm(stdForm string) (PhoneNumber, error) {
	var num PhoneNumber
	segmentsSlice := strings.Split(stdForm, sep)
	if len(segmentsSlice) != 2 {
		return num, apperr.ErrInvalidPhoneNumber
	}

	num.CountryCode = segmentsSlice[0]
	num.Number = segmentsSlice[1]

	return num, nil
}

func (p PhoneNumber) ToString() string {
	return p.CountryCode + p.Number
}

// 963|123456789
//
// DON't change this form befor you migrate your database to the new form
// change it will cause the old stored phone numbers to not be read correctly
func (p PhoneNumber) ToAppStdForm() string {
	return fmt.Sprintf("%s%s%s", p.CountryCode, sep, p.Number)
}

func (p PhoneNumber) IsValid() bool {
	// TODO: implement this somehow!!
	return len(p.CountryCode) != 0 && len(p.Number) != 0
}
