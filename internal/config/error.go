package config

import (
	"errors"
)

var (
	ErrUnimplemented = errors.New("not implemented error")

	ErrMalformed = errors.New("malformation error")
)

type ConfigError struct {
	err     error
	message string
}

func (e *ConfigError) Error() string {
	return e.err.Error() + ": " + e.message
}

func (e *ConfigError) Unwrap() error {
	return e.err
}
