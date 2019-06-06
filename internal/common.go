package internal

import (
	"errors"
	"flag"
)

// Verbose used to get the verbose logs
var Verbose = flag.Bool("verbose-api", false, "Used to get verbose data")

// ScopeIDCtx is the common struct used as a scope's ID in the context
type ScopeIDCtx struct{}

// ControllerKeyCtx is the common struct used as a controller's key in the context
type ControllerKeyCtx struct{}

// TransactionStateCtx is the common struct used as a transaction state key in the scope's context
type TransactionStateCtx struct{}

// ReducedPrimariesCtx is the context key that keeps the primary values in the patch scope
type ReducedPrimariesCtx struct{}

// PreviousProcess is a struct used as the context key for getting the previous process
type PreviousProcess struct{}

// Context
var (
	ScopeIDCtxKey          = ScopeIDCtx{}
	ControllerCtxKey       = ControllerKeyCtx{}
	TxStateCtxKey          = TransactionStateCtx{}
	ReducedPrimariesCtxKey = ReducedPrimariesCtx{}
	PreviousProcessCtxKey  = PreviousProcess{}
)

// Errors used in by the internal packages
var (
	ErrUnexpectedType     = errors.New("models should be a struct pointer or slice of struct pointers")
	ErrExpectedSlice      = errors.New("models should be a slice of struct pointers")
	ErrNilValue           = errors.New("Nil value provided.")
	ErrValueNotAddresable = errors.New("Provided value is not addressable")

	// Unmarshal Errors
	ErrClientIDDisallowed = errors.New("Client id is disallowed for this model.")
	// ErrInvalidTime is returned when a struct has a time.Time type field, but
	// the JSON value was not a unix timestamp integer.
	ErrInvalidTime = errors.New("Only numbers can be parsed as dates, unix timestamps")
	// ErrInvalidISO8601 is returned when a struct has a time.Time type field and includes
	// "iso8601" in the tag spec, but the JSON value was not an ISO8601 timestamp string.
	ErrInvalidISO8601 = errors.New("Only strings can be parsed as dates, ISO8601 timestamps")
	// ErrUnknownFieldNumberType is returned when the JSON value was a float
	// (numeric) but the Struct field was a non numeric type (i.e. not int, uint,
	// float, etc)
	ErrUnknownFieldNumberType = errors.New("The struct field was not of a known number type")
	// ErrUnsupportedPtrType is returned when the Struct field was a pointer but
	// the JSON value was of a different type
	ErrUnsupportedPtrType = errors.New("Pointer type in struct is not supported")
	// ErrInvalidType is returned when the given type is incompatible with the expected type.
	ErrInvalidType = errors.New("Invalid type provided") // I wish we used punctuation.
	// ErrInvalidQuery

	// ErrBadJSONAPIID is returned when the Struct JSON API annotated "id" field
	// was not a valid numeric type.
	ErrBadJSONAPIID = errors.New(
		"id should be either string, int(8,16,32,64) or uint(8,16,32,64)")

	ErrModelNotMapped = errors.New("Unmapped model provided.")

	ErrFieldNotFound        = errors.New("Field not found")
	ErrFieldAlreadySelected = errors.New("Field already selected.")
)
