package service

import (
	"errors"
	"fmt"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
)

type Error struct {
	Kind    error
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}

	if e.Err != nil {
		return e.Err.Error()
	}

	return e.Kind.Error()
}

func (e *Error) Unwrap() error {
	if e.Err != nil {
		return e.Err
	}

	return e.Kind
}

func (e *Error) Is(target error) bool {
	return errors.Is(e.Kind, target)
}

func validationError(message string) error {
	return &Error{
		Kind:    ErrValidation,
		Message: message,
	}
}

func notFoundError(message string) error {
	return &Error{
		Kind:    ErrNotFound,
		Message: message,
	}
}

func wrapError(message string, err error) error {
	return fmt.Errorf("%s: %w", message, err)
}
