package emailvalidator

import (
	"net/mail"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
)

func IsValidEmail(email string) bool {
	a, err := mail.ParseAddress(email)
	return err == nil && a.Address == email
}

func IsValidEmailErr(email string) error {
	ok := IsValidEmail(email)
	if !ok {
		return apperr.ErrInvalidEmail
	}
	return nil
}
