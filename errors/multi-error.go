package errors

import (
	"errors"
	"strings"
)

// MultiError is the slice of errors parsable into a single error.
type MultiError []error

// Is checks if one of the errors is equal to 'err'.
func (m MultiError) Is(err error) bool {
	for _, subError := range m {
		if errors.Is(subError, err) {
			return true
		}
	}
	return false
}

// Error implements error interface.
func (m MultiError) Error() string {
	sb := &strings.Builder{}

	for i, e := range m {
		sb.WriteString(e.Error())
		if i != len(m)-1 {
			sb.WriteString(",")
		}
	}
	return sb.String()
}
