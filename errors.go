package jsonapi

import (
	"fmt"
)

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

// AddMeta adds the meta data for given error. Checks if an object has inited meta field.
func (e *ErrorObject) AddMeta(key string, value interface{}) {
	if e.Meta == nil {
		*e.Meta = (make(map[string]interface{}))
	}
	meta := *e.Meta
	meta[key] = value
	return
}
