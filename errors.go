package jsonapi

import (
	"encoding/json"
	"fmt"
	"io"
)

// MarshalErrors writes a JSON API response using the given `[]error`.
//
// For more information on JSON API error payloads, see the spec here:
// http://jsonapi.org/format/#document-top-level
// and here: http://jsonapi.org/format/#error-objects.
func MarshalErrors(w io.Writer, errorObjects ...*ErrorObject) error {
	if err := json.NewEncoder(w).Encode(&ErrorsPayload{Errors: errorObjects}); err != nil {
		return err
	}
	return nil
}

// ErrorsPayload is a serializer struct for representing a valid JSON API errors payload.
type ErrorsPayload struct {
	Errors []*ErrorObject `json:"errors"`
}

// ErrorObject is an struct representing a JSON API error
type ErrorObject struct {
	// ID is a unique identifier for this particular occurrence of a problem.
	ID string `json:"id,omitempty"`

	// Title is a short, human-readable summary of the problem that SHOULD NOT change from occurrence to occurrence of the problem, except for purposes of localization.
	Title string `json:"title,omitempty"`

	// Detail is a human-readable explanation specific to this occurrence of the problem. Like title, this field’s value can be localized.
	Detail string `json:"detail,omitempty"`

	// Status is the HTTP status code applicable to this problem, expressed as a string value.
	Status string `json:"status,omitempty"`

	// Code is an application-specific error code, expressed as a string value.
	Code string `json:"code,omitempty"`

	// Meta is an object containing non-standard meta-information about the error.
	Meta *map[string]interface{} `json:"meta,omitempty"`

	// Err is a non published message container for loggin purpose
	Err error `json:"-"`
}

// Copy returns the new object that is a copy of given error object.
func (e ErrorObject) Copy() *ErrorObject {
	err := e
	return &err
}

// Error implements the `Error` interface.
func (e *ErrorObject) Error() string {
	return fmt.Sprintf("Error: %s %s\n", e.Title, e.Detail)
}

// WithDetail sets the detail for given error and then returns the error.
func (e *ErrorObject) WithDetail(detail string) *ErrorObject {
	e.Detail = detail
	return e
}

// AddMeta adds the meta data for given error. Checks if an object has inited meta field.
func (e *ErrorObject) AddMeta(key string, value interface{}) {
	if e.Meta == nil {
		mp := make(map[string]interface{})
		e.Meta = &mp
	}
	meta := *e.Meta
	meta[key] = value
	return
}

type ErrorCode int

const (
	HErrBadValues ErrorCode = iota
	HErrNoValues
	HErrNoModel
	HErrAlreadyWritten
	HErrInternal
	HErrValuePreset
	HErrWarning
)

type HandlerError struct {
	Code    ErrorCode
	Message string
	Scope   *Scope
	Field   *StructField
	Model   *ModelStruct
}

func newHandlerError(code ErrorCode, msg string) *HandlerError {
	return &HandlerError{Code: code, Message: msg}
}

func (e *HandlerError) Error() string {
	return fmt.Sprintf("%d. %s", e.Code, e.Message)
}
