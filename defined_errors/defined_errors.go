package definederrors

import "errors"

var (
	ErrorInput           = errors.New("input was invalid")
	ErrorNoArgs          = errors.New("no arguments were passed in")
	ErrorWrongNumArgs    = errors.New("wrong number of arguments passed in ")
	ErrorHandlerNotExist = errors.New("handler does not exist")
	ErrorNilPointer      = errors.New("pointer is nil")
)
