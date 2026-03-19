package service

import "errors"

var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInvalidArgument = errors.New("invalid_argument")
)
