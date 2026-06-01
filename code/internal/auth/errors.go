package auth

import "errors"

var (
	ErrEmailTaken        = errors.New("email already in use")
	ErrPasswordTooShort  = errors.New("password must be at least 8 characters")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked     = errors.New("account is temporarily locked")
	ErrNotFound          = errors.New("user not found")
	ErrTokenExpired      = errors.New("token has expired")
	ErrTokenInvalid      = errors.New("token is invalid")
	ErrNotVerified       = errors.New("email not verified")
	ErrCannotBlockSelf   = errors.New("cannot block yourself")
)
