package utils

import (
	"errors"
	"fmt"
	"math"
	"net/mail"

	"github.com/Nidal-Bakir/go-todo-backend/internal/appenv"
)

var (
	ErrNotValidEmail = errors.New("text string")
)

func SafeIntToInt32(v int) (int32, error) {
	if v < math.MinInt32 || v > math.MaxInt32 {
		return 0, fmt.Errorf("value %d out of range for int32", v)
	}
	return int32(v), nil
}

func Assert(ok bool, v any) {
	if !ok {
		panic(v)
	}
}

func AssertDev(ok bool, v any) {
	if !ok && appenv.IsStagOrLocal() {
		panic(v)
	}
}

func AssertDevFn(v any, fn func() bool) {
	if appenv.IsStagOrLocal() {
		if !fn() {
			panic(v)
		}
	}
}

// usage e.g:
//
//	func success() (int, error) {
//		return 0, nil
//	}
//	n1 := Must(success())
func Must[T any](d T, err error) T {
	if err != nil {
		panic(err)
	}
	return d
}

// usage e.g:
//
//	func success() (int, string, error) {
//		return 0, "hi", nil
//	}
//	d1, d2 := Must2(success())
func Must2[T1 any, T2 any](d1 T1, d2 T2, err error) (T1, T2) {
	if err != nil {
		panic(err)
	}
	return d1, d2
}

func IsValidEmail(email string) bool {
	a, err := mail.ParseAddress(email)
	return err == nil && a.Address == email
}

func IsValidEmailErr(email string) error {
	ok := IsValidEmail(email)
	if !ok {
		return ErrNotValidEmail
	}
	return nil
}
