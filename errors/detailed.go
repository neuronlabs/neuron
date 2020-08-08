package errors

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/google/uuid"
)

// compile time check for DetailedError interfaces.
var (
	_ error = &DetailedError{}
)

// DetailedError is the class based error definition.
// Each instance has it's own traceable ID.
// It contains also a Class variable that might be comparable in logic.
type DetailedError struct {
	// ID is a unique error instance identification number.
	ID uuid.UUID
	// details contains the detailed information.
	Details string
	// message is a message used as a string for the
	// golang error interface implementation.
	Message string
	// Operation is the operation name when the error occurred.
	Operation string

	err error
}

// WithDetail sets error detail.
func (e *DetailedError) WithDetail(detail string) *DetailedError {
	e.Details = detail
	return e
}

// WithDetailf sets error detail.
func (e *DetailedError) WithDetailf(format string, values ...interface{}) *DetailedError {
	e.Details = fmt.Sprintf(format, values...)
	return e
}

// WrapDet wraps 'err' error and creates DetailedError with given 'message'.
func WrapDet(err error, message string) *DetailedError {
	detailedErr := newDetailed(err)
	detailedErr.Message = message
	return detailedErr
}

// WrapDetf wraps 'err' error and created DetailedError with formatted message.
func WrapDetf(err error, format string, args ...interface{}) *DetailedError {
	detailedErr := newDetailed(err)
	detailedErr.Message = fmt.Sprintf(format, args...)
	return detailedErr
}

// DetailedError implements error interface.
func (e *DetailedError) Error() string {
	return e.Message
}

// Unwrap unwraps provided error.
func (e *DetailedError) Unwrap() error {
	return e.err
}

func newDetailed(err error) *DetailedError {
	detailedErr := &DetailedError{
		ID:  uuid.New(),
		err: err,
	}
	pc, _, _, ok := runtime.Caller(2)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		file, line := details.FileLine(pc)
		_, singleFile := filepath.Split(file)
		detailedErr.Operation = details.Name() + "#" + singleFile + ":" + strconv.Itoa(line)
	}
	return detailedErr
}
