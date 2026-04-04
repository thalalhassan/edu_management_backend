package constants

import "errors"

type ErrorConstants struct {
	ErrUserNotFound       error
	ErrInvalidCredentials error
	ErrUserExists         error
}

var Errors = ErrorConstants{
	ErrUserNotFound:       errors.New("user not found"),
	ErrInvalidCredentials: errors.New("invalid credentials"),
	ErrUserExists:         errors.New("user already exists"),
}
