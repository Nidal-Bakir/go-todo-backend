package middleware

import (
	"reflect"
	"testing"

	_ "github.com/Nidal-Bakir/go-todo-backend/testing_init"
)

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expecting no error")
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expecting error")
	}
}

func assertEqual(t *testing.T, a, b interface{}) {
	t.Helper()
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("expecting values to be equal but got: '%v' and '%v'", a, b)
	}
}
