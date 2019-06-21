package errors

import (
	"strings"

	"github.com/neuronlabs/neuron/errors/class"
)

// MultiError is the slice of errors parsable into a single error.
type MultiError []*Error

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

// HasMajor checks if provided 'major' occurs in given multi error slice.
func (m MultiError) HasMajor(major class.Major) bool {
	for _, err := range m {
		if err.Class.IsMajor(major) {
			return true
		}
	}
	return false
}
