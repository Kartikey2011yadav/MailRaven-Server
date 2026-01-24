package ports

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrAlreadyExists      = errors.New("resource already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrStorageFailure     = errors.New("storage operation failed")
)
