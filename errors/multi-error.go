package errors

import (
	"strings"
)

// MultiError is the slice of errors parsable into a single error.
type MultiError []error

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
