// Package gcerrors provides error handling for the gocache package.
package gcerrors

import "fmt"

// Error is an error that contains a scheme and an error.
type Error struct {
	scheme string
	err    error
}

// Error returns the error message.
func (e *Error) Error() string {
	if e.scheme == "" {
		return fmt.Sprintf("gocache: %s", e.err.Error())
	} else {
		return fmt.Sprintf("gocache/%s: %s", e.scheme, e.err.Error())
	}
}

// Unwrap returns the original error.
func (e *Error) Unwrap() error {
	return e.err
}

// NewWithScheme returns a new error with the given scheme and error.
func NewWithScheme(scheme string, err error) *Error {
	return &Error{scheme: scheme, err: err}
}

// New returns a new error with the given error.
func New(err error) *Error {
	return &Error{err: err}
}
