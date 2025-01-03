package errorutils

import "errors"

func CheckErrTypeMatch(got, expected error) bool {
	return errors.Is(got, expected)
}
