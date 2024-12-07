package definedErrors

import "errors"

var (
	ErrLoginHandlerZeroArgs = errors.New("the login handler expects a single argument, the username")
	ErrUserNameNil          = errors.New("username cannot be nil")
	ErrCommandMapNil        = errors.New("the command map in commands is a nil map")
	ErrCommandNotExist      = errors.New("the command has not yet been registered")
	ErrNotEnoughArguments   = errors.New("not enough arguments were provided")
)
