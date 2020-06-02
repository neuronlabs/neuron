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
	_ ClassError = &DetailedError{}
)

// DetailedError is the class based error definition.
// Each instance has it's own trackable ID. It's chainable
// It contains also a Class variable that might be comparable in logic.
type DetailedError struct {
	// ID is a unique error instance identification number.
	ID uuid.UUID
	// Classification defines the error classification.
	Classification Class
	// details contains the detailed information.
	Details string
	// message is a message used as a string for the
	// golang error interface implementation.
	Message string
	// Opertaion is the operation name when the error occurred.
	Operation string
}

// NewDet creates DetailedError with given 'class' and message 'message'.
func NewDet(c Class, message string) *DetailedError {
	err := newDetailed(c)
	err.Message = message
	return err
}

// NewDetf creates DetailedError instance with provided 'class' with formatted message.
// DetailedError implements ClassError interface.
func NewDetf(c Class, format string, args ...interface{}) *DetailedError {
	err := newDetailed(c)
	err.Message = fmt.Sprintf(format, args...)
	return err
}

// Class implements ClassError.
func (e *DetailedError) Class() Class {
	return e.Classification
}

// DetailedError implements error interface.
func (e *DetailedError) Error() string {
	return e.Message
}

func newDetailed(c Class) *DetailedError {
	err := &DetailedError{
		ID:             uuid.New(),
		Classification: c,
	}
	pc, _, _, ok := runtime.Caller(2)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		file, line := details.FileLine(pc)
		_, singleFile := filepath.Split(file)
		err.Operation = details.Name() + "#" + singleFile + ":" + strconv.Itoa(line)
	}
	return err
}
