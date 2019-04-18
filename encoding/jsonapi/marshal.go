package jsonapi

import (
	"encoding/json"
	ctrl "github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	ictrl "github.com/neuronlabs/neuron/internal/controller"
	"io"
)

// Marshal marshals the provided value 'v' into the writer
func Marshal(w io.Writer, v interface{}) error {
	return (*ictrl.Controller)(ctrl.Default()).Marshal(w, v)
}

// MarshalC marshals the provided value 'v' into the writer. It uses the 'c' controller
func MarshalC(c *ctrl.Controller, w io.Writer, v interface{}) error {
	return (*ictrl.Controller)(c).Marshal(w, v)
}

// MarshalErrors writes a JSON API response using the given `[]error`.
//
// For more information on JSON API error payloads, see the spec here:
// http://jsonapi.org/format/#document-top-level
// and here: http://jsonapi.org/format/#error-objects.
func MarshalErrors(w io.Writer, errorObjects ...*errors.ApiError) error {
	if err := json.NewEncoder(w).Encode(&ErrorsPayload{Errors: errorObjects}); err != nil {
		return err
	}
	return nil
}

// ErrorsPayload is a serializer struct for representing a valid JSON API errors payload.
type ErrorsPayload struct {
	Errors []*errors.ApiError `json:"errors"`
}
