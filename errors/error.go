package errors

import (
	"errors"
	"fmt"
)

// ErrInternal is an internal neuron error definition.
var ErrInternal = errors.New("internal error")

var _ error = &simpleError{}

type simpleError struct {
	err error
	msg string
}

// New creates new error.
func New(msg string) error {
	return errors.New(msg)
}

// Wrap creates simple ClassError for provided 'c' Class and 'msg' message.
func Wrap(err error, msg string) error {
	return &simpleError{err, msg}
}

// Wrapf creates simple formatted ClassError for provided 'c' Class, 'format' and arguments 'args'.
func Wrapf(err error, format string, args ...interface{}) error {
	return &simpleError{err, fmt.Sprintf(format, args...)}
}

// Error implements error interface.
func (s *simpleError) Error() string {
	return s.msg
}

// Unwrap unwraps provided error.
func (s *simpleError) Unwrap() error {
	return s.err
}

// Is checks if any of the chained errors is of 'target' type.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As executes errors.As function.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Unwrap unwraps provided error.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}
