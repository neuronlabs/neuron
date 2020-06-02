package codec

import (
	"strings"

	"github.com/neuronlabs/neuron/errors"
)

var (
	MjrCodec errors.Major

	ClassMarshal   errors.Class
	ClassUnmarshal errors.Class
	ClassNoModels  errors.Class
)

func init() {
	MjrCodec = errors.MustNewMajor()

	ClassMarshal = errors.MustNewMajorClass(MjrCodec)
	ClassUnmarshal = errors.MustNewMajorClass(MjrCodec)
	ClassNoModels = errors.MustNewMajorClass(MjrCodec)
}

// Error is the error structure used for used for codec processes.
// More info can be found at: 'https://jsonapi.org/format/#errors'
type Error struct {
	// ID is a unique identifier for this particular occurrence of a problem.
	ID string `json:"id,omitempty"`
	// Title is a short, human-readable summary of the problem that SHOULD NOT change from occurrence to occurrence of the problem, except for purposes of localization.
	Title string `json:"title,omitempty"`
	// Detail is a human-readable explanation specific to this occurrence of the problem. Like title, this field’s value can be localized.
	Detail string `json:"detail,omitempty"`
	// Status is the status code applicable to this problem, expressed as a string value.
	Status string `json:"status,omitempty"`
	// Code is an application-specific error code, expressed as a string value.
	Code string `json:"code,omitempty"`
	// Meta is an object containing non-standard meta-information about the error.
	Meta map[string]interface{} `json:"meta,omitempty"`
}

func (e *Error) Error() string {
	sb := strings.Builder{}
	e.error(&sb)
	return sb.String()
}

func (e *Error) error(sb *strings.Builder) {
	if e.ID != "" {
		sb.WriteString(e.ID)
		sb.WriteString(" - ")
	}
	if e.Status != "" {
		sb.WriteRune('[')
		sb.WriteString(e.Status)
		sb.WriteRune(']')
		sb.WriteRune(' ')
	}
	if e.Title != "" {
		sb.WriteString(e.Title)
		sb.WriteRune(' ')
	}
	if e.Detail != "" {
		sb.WriteString(e.Detail)
		sb.WriteRune(' ')
	}
	if e.Code != "" {
		sb.WriteString("CODE: ")
		sb.WriteString(e.Code)
	}
}

type MultiError []*Error

func (m MultiError) Error() string {
	sb := &strings.Builder{}
	for i, err := range m {
		err.error(sb)
		if i != len(m)-1 {
			sb.WriteString(" | ")
		}
	}
	return sb.String()
}
