package errors

import (
	"fmt"
)

// ApiError is an struct representing a JSON API error
type ApiError struct {
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
func (e ApiError) Copy() *ApiError {
	err := e
	return &err
}

// ApiError implements the `ApiError` interface.
func (e *ApiError) Error() string {
	return fmt.Sprintf("ApiError: %s %s\n", e.Title, e.Detail)
}

// WithDetail sets the detail for given error and then returns the error.
func (e *ApiError) WithDetail(detail string) *ApiError {
	e.Detail = detail
	return e
}

// AddMeta adds the meta data for given error. Checks if an object has inited meta field.
func (e *ApiError) AddMeta(key string, value interface{}) {
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
