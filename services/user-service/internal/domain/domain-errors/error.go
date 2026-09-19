package domainErrors

import "errors"

var (
	ErrInvalidUserID          = errors.New("invalid user id")
	ErrInvalidEmail           = errors.New("invalid email")
	ErrInvalidRole            = errors.New("invalid role")
	ErrInvalidPasswordHash    = errors.New("invalid password hash")
	ErrUserAlreadyInactive    = errors.New("user is already inactive")
	ErrUserAlreadyActive      = errors.New("user is already active")
	ErrRoleUnchanged          = errors.New("user already has this role")
	ErrPasswordUnchanged      = errors.New("new password matches current password")
	ErrCorruptedUser          = errors.New("user is corrupted")
	ErrUserNotFound           = errors.New("user not found")
	ErrEmailAlreadyRegistered = errors.New("email already registered")
)
