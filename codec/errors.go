package codec

import (
	"strconv"
	"strings"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
)

var (
	// ErrCodec is a general codec error.
	ErrCodec = errors.New("codec")
	// ErrMarshal is an error with marshaling.
	ErrMarshal = errors.Wrap(ErrCodec, "marshal")
	// ErrMarshalPayload is an error of marshaling payload.
	ErrMarshalPayload = errors.Wrap(ErrMarshal, "payload")

	// ErrUnmarshal is an error related to unmarshaling process.
	ErrUnmarshal = errors.Wrap(ErrCodec, "unmarshal")
	// ErrUnmarshalDocument is an error related to unmarshaling invalid document,
	ErrUnmarshalDocument = errors.Wrap(ErrUnmarshal, "document")
	// ErrUnmarshalFieldValue is an error related to unmarshaling field value.
	ErrUnmarshalFieldValue = errors.Wrap(ErrUnmarshal, "field value")
	// ErrUnmarshalFieldName is an error related t unmarshaling field name
	ErrUnmarshalFieldName = errors.Wrap(ErrUnmarshal, "field name")

	// ErrOptions is an error that defines invalid marshal/unmarshal options.
	ErrOptions = errors.Wrap(ErrCodec, "options")
)

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
	Meta Meta `json:"meta,omitempty"`
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

// MultiError is an error composed of multiple single errors.
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

// Status gets the most significant api error status.
func (m MultiError) Status() int {
	var highestStatus int
	for _, err := range m {
		status, er := strconv.Atoi(err.Status)
		if er != nil {
			log.Warningf("Error: '%v' contains non integer status value", err)
			continue
		}
		if err.Status == "500" {
			return 500
		}
		if status > highestStatus {
			highestStatus = status
		}
	}
	if highestStatus == 0 {
		highestStatus = 500
	}
	return highestStatus
}
