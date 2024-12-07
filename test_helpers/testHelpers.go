package testHelpers

import "testing"

func AssertStrings(got, want string, t testing.TB) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot: %s\nwant: %s\n", got, want)
	}
}

func AssertNoError(err error, t testing.TB) {
	if err != nil {
		t.Errorf("didn't expect error, got %v", err)
	}
}

func AssertHasError(err error, t testing.TB) {
	if err == nil {
		t.Errorf("expected error, didn't get one")
	}
}
