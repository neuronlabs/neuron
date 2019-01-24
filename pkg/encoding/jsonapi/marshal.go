package jsonapi

import (
	// ictrl "github.com/kucjac/jsonapi/pkg/internal/controller"
	// ctrl "github.com/kucjac/jsonapi/pkg/controller"
	"encoding/json"
	"github.com/kucjac/jsonapi/pkg/errors"
	"io"
)

func Marshal(w io.Writer, v interface{}) error {
	return nil
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
