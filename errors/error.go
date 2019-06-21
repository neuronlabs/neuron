package errors

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/neuronlabs/neuron/errors/class"
)

// Error is the common error definition used in the neuron project.
type Error struct {
	// ID is a unique error instance identification number.
	ID uuid.UUID

	// Class defines the error classification.
	Class class.Class

	// Detail contains the detailed information.
	Detail string

	// InternalMessage is a message used as a string for the
	// golang error interface implementation.
	InternalMessage string

	// Opertaion is the operation name when the error occurred.
	Operation string
}

// Error implements error interface.
func (e *Error) Error() string {
	return e.InternalMessage
}

// SetClass sets the error Class 'c'  and returns itself.
func (e *Error) SetClass(c class.Class) *Error {
	e.Class = c
	return e
}

// SetDetail sets the error 'detail' and returns itself.
func (e *Error) SetDetail(detail string) *Error {
	e.Detail = detail
	return e
}

// SetDetailf sets the error's formatted detail with provided and returns itself.
func (e *Error) SetDetailf(format string, args ...interface{}) *Error {
	e.Detail = fmt.Sprintf(format, args...)
	return e
}

// WrapDetail wraps the 'detail' for given error. Wrapping appends the new detail
// to the front of error detail message.
func (e *Error) WrapDetail(detail string) *Error {
	return e.wrapDetail(detail)
}

// WrapDetailf wraps the detail with provided formatting for given error.
// Wrapping appends the new detail to the front of error detail message.
func (e *Error) WrapDetailf(format string, args ...interface{}) *Error {
	return e.wrapDetail(fmt.Sprintf(format, args...))
}

// SetOperation sets the error's operation and returns error by itself.
func (e *Error) SetOperation(operation string) *Error {
	e.Operation = operation
	return e
}

func (e *Error) wrapDetail(detail string) *Error {
	if e.Detail == "" {
		e.Detail = detail
	} else {
		e.Detail = detail + " " + e.Detail
	}
	return e
}

// New creates new error message with given 'class' and message 'message'.
func New(c class.Class, message string) *Error {
	return &Error{
		ID:              uuid.New(),
		Class:           c,
		InternalMessage: message,
	}
}

// Newf creates new error instance with provided 'class' with formatted message.
func Newf(c class.Class, format string, args ...interface{}) *Error {
	return &Error{
		ID:              uuid.New(),
		Class:           c,
		InternalMessage: fmt.Sprintf(format, args...),
	}
}
