package testHelpers

import (
	"errors"
	"testing"
)

func AssertStrings(got, want string, t testing.TB) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot: %s\nwant: %s\n", got, want)
	}
}

func AssertNoError(err error, t testing.TB) {
	t.Helper()
	if err != nil {
		t.Errorf("didn't expect error, got %v", err)
	}
}

func AssertHasError(err error, t testing.TB) {
	t.Helper()
	if err == nil {
		t.Errorf("expected error, didn't get one")
	}
}

func AssertErrorType(got, want error, t testing.TB) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Errorf("\nexpect error %v, got error %v\n", got, want)
	}
}
