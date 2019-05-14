package internal

import (
	"errors"
	"flag"
)

var Verbose *bool = flag.Bool("verbose-api", false, "Used to get verbose data")

// ScopeIDCtx is the common struct used as a scope's ID in the context
type ScopeIDCtx struct{}

// ControllerIDCtx is the common struct used as a controller's key in the context
type ControllerIDCtx struct{}

var (
	ScopeIDCtxKey      = ScopeIDCtx{}
	ControllerIDCtxKey = ControllerIDCtx{}
)

// Erros used in by the internal packages
var (
	IErrUnexpectedType     = errors.New("models should be a struct pointer or slice of struct pointers")
	IErrExpectedSlice      = errors.New("models should be a slice of struct pointers")
	IErrNilValue           = errors.New("Nil value provided.")
	IErrValueNotAddresable = errors.New("Provided value is not addressable")

	// Unmarshal Errors
	IErrClientIDDisallowed = errors.New("Client id is disallowed for this model.")
	// IErrInvalidTime is returned when a struct has a time.Time type field, but
	// the JSON value was not a unix timestamp integer.
	IErrInvalidTime = errors.New("Only numbers can be parsed as dates, unix timestamps")
	// IErrInvalidISO8601 is returned when a struct has a time.Time type field and includes
	// "iso8601" in the tag spec, but the JSON value was not an ISO8601 timestamp string.
	IErrInvalidISO8601 = errors.New("Only strings can be parsed as dates, ISO8601 timestamps")
	// IErrUnknownFieldNumberType is returned when the JSON value was a float
	// (numeric) but the Struct field was a non numeric type (i.e. not int, uint,
	// float, etc)
	IErrUnknownFieldNumberType = errors.New("The struct field was not of a known number type")
	// IErrUnsupportedPtrType is returned when the Struct field was a pointer but
	// the JSON value was of a different type
	IErrUnsupportedPtrType = errors.New("Pointer type in struct is not supported")
	// IErrInvalidType is returned when the given type is incompatible with the expected type.
	IErrInvalidType = errors.New("Invalid type provided") // I wish we used punctuation.
	// ErrInvalidQuery

	// IErrBadJSONAPIStructTag is returned when the Struct field's JSON API
	// annotation is invalid.
	IErrBadJSONAPIStructTag = errors.New("Bad jsonapi struct tag format")
	// IErrBadJSONAPIID is returned when the Struct JSON API annotated "id" field
	// was not a valid numeric type.
	IErrBadJSONAPIID = errors.New(
		"id should be either string, int(8,16,32,64) or uint(8,16,32,64)")

	IErrModelNotMapped = errors.New("Unmapped model provided.")

	IErrFieldNotFound        = errors.New("Field not found")
	IErrFieldAlreadySelected = errors.New("Field already selected.")
)
