package service

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrInvalid occurs when checks fails
	ErrInvalid = errors.New("invalid")

	// ErrUnauthorized occurs when user is not authorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden occurs when user if forbideen
	ErrForbidden = errors.New("forbidden")

	// ErrNotFound occurs when somehint is not found
	ErrNotFound = errors.New("not found")

	// ErrInternalError occurs when shit happens
	ErrInternalError = errors.New("internal error")
)

// ConcatError concat errors to a single string
func ConcatError(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	values := make([]string, len(errs))
	for index, err := range errs {
		values[index] = err.Error()
	}

	return errors.New(strings.Join(values, ", "))
}

// WrapInvalid wraps given error with invalid err
func WrapInvalid(err error) error {
	return fmt.Errorf("%s: %w", err, ErrInvalid)
}

// WrapUnauthorized wraps given error with unauthorized err
func WrapUnauthorized(err error) error {
	return fmt.Errorf("%s: %w", err, ErrUnauthorized)
}

// WrapForbidden wraps given error with forbidden err
func WrapForbidden(err error) error {
	return fmt.Errorf("%s: %w", err, ErrForbidden)
}

// WrapInternal wraps given error with internal err
func WrapInternal(err error) error {
	return fmt.Errorf("%s: %w", err, ErrInternalError)
}

// WrapNotFound wraps given error with not found err
func WrapNotFound(err error) error {
	return fmt.Errorf("%s: %w", err, ErrNotFound)
}
