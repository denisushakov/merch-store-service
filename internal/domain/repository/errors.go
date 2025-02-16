package repository

import (
	"errors"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUnauthorized = errors.New("user unauthorized")
)
