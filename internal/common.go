package internal

// ControllerStoreKeyStruct is the common struct used as a controller's key in the Stores
type ControllerStoreKeyStruct struct{}

// TxStateStoreStruct is the common struct used as a transaction state key in the scope's Store
type TxStateStoreStruct struct{}

// ReducedPrimariesStoreStruct is the Store key that keeps the primary values in the patch scope
type ReducedPrimariesStoreStruct struct{}

type primariesChecked struct{}

// PreviousProcess is a struct used as the Store key for getting the previous process
type PreviousProcess struct{}

type autoBegin struct{}

type created struct{}

// Store keys
var (
	// ControllerStoreKey is the store key used to store the controller.
	ControllerStoreKey = ControllerStoreKeyStruct{}

	//  TxStateStoreKey is the store key used to store the scope's transaction.
	TxStateStoreKey = TxStateStoreStruct{}

	// ReducedPrimariesStoreKey is the store key used to store the reduced scope primary values.
	ReducedPrimariesStoreKey = ReducedPrimariesStoreStruct{}

	// PrimariesAlreadyChecked is the store key flag value used to notify that the primaries for given process chain were already checked.
	PrimariesAlreadyChecked = primariesChecked{}

	// PreviousProcessStoreKey is the store key used to indicate what was the previous query process.
	PreviousProcessStoreKey = PreviousProcess{}

	// AutoBeginStoreKey is the key that defines if query were started with begin_transaction process.
	AutoBeginStoreKey = autoBegin{}

	// JustCreated is the store key that defines that the root scope's process is just after the creation process.
	JustCreated = created{}
)

// TODO: clear the error definitions below after mapping the gateway errors.
// Errors used in by the internal packages

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
