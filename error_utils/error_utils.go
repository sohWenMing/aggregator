package errorutils

import (
	"errors"

	"github.com/lib/pq"
)

func CheckErrTypeMatch(got, expected error) bool {
	return errors.Is(got, expected)
}

func UnwrapPqErr(err error) (isPQErr bool, pqErr *pq.Error, rawErr error) {

	processedErr := err
	for processedErr != nil {
		//checks if the current error being proccesed is an instance of pq.Error, if so, returns the current error
		if postgresErr, ok := processedErr.(*pq.Error); ok {
			return true, postgresErr, err
		}
		processedErr = errors.Unwrap(processedErr)
	}
	return false, nil, err
}
