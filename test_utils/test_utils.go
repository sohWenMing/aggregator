package testutils

import (
	"testing"
)

func AssertStrings(got, want string, t testing.TB) {
	if got != want {
		t.Errorf("got: %q\nwant: %q\n", got, want)
	}
}

func AssertInts(got, want int, t testing.TB) {
	if got != want {
		t.Errorf("got: %d\nwant: %d\n", got, want)
	}
}

func AssertNoErr(err error, t testing.TB) {
	if err != nil {
		t.Errorf("didn't expect error, got %v\n", err)
	}
}

func AssertHasErr(err error, t testing.TB) {
	if err == nil {
		t.Errorf("expected error, didn't get one\n")
	}
}
