package database

import (
	"github.com/lib/pq"
	errorutils "github.com/sohWenMing/aggregator/error_utils"
)

func CheckPqErr(err error) (isPQErr, isUniqueViolation bool,
	pqErr *pq.Error, rawErr error) {
	isPQErr, pqErr, rawErr = errorutils.UnwrapPqErr(err)
	if isPQErr {
		if pqErr.Code == "23505" {
			return true, true, pqErr, rawErr
		}
		return true, false, pqErr, rawErr
	}
	return false, false, nil, rawErr

}
