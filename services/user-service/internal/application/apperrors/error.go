package apperrors

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountDeactivated = errors.New("account is deactivated")
)
