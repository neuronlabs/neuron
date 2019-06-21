package internal

// ControllerStoreKeyStruct is the common struct used as a controller's key in the Stores
type ControllerStoreKeyStruct struct{}

// TxStateStoreStruct is the common struct used as a transaction state key in the scope's Store
type TxStateStoreStruct struct{}

// ReducedPrimariesStoreStruct is the Store key that keeps the primary values in the patch scope
type ReducedPrimariesStoreStruct struct{}

// PreviousProcess is a struct used as the Store key for getting the previous process
type PreviousProcess struct{}

// Store keys
var (
	ControllerStoreKey       = ControllerStoreKeyStruct{}
	TxStateStoreKey          = TxStateStoreStruct{}
	ReducedPrimariesStoreKey = ReducedPrimariesStoreStruct{}
	PreviousProcessStoreKey  = PreviousProcess{}
)

// TODO: clear the errors
// Errors used in by the internal packages
var (

// Unmarshal Errors

// ErrInvalidTime is returned when a struct has a time.Time type field, but
// the JSON value was not a unix timestamp integer.

// ErrInvalidISO8601 is returned when a struct has a time.Time type field and includes
// "iso8601" in the tag spec, but the JSON value was not an ISO8601 timestamp string.

// ErrUnknownFieldNumberType is returned when the JSON value was a float
// (numeric) but the Struct field was a non numeric type (i.e. not int, uint,
// float, etc)

// ErrUnsupportedPtrType is returned when the Struct field was a pointer but
// the JSON value was of a different type

// ErrInvalidQuery

// ErrBadJSONAPIID is returned when the Struct JSON API annotated "id" field
// was not a valid numeric type.

)
