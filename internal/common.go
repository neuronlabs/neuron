package internal

// ControllerStoreKeyStruct is the common struct used as a controller's key in the Stores.
type ControllerStoreKeyStruct struct{}

// TxStateStoreStruct is the common struct used as a transaction state key in the scope's Store.
type TxStateStoreStruct struct{}

// ReducedPrimariesStoreStruct is the Store key that keeps the primary values in the patch scope.
type ReducedPrimariesStoreStruct struct{}

type primariesChecked struct{}

// PreviousProcess is a struct used as the Store key for getting the previous process.
type PreviousProcess struct{}

type autoBegin struct{}

type created struct{}

// Store keys
var (
	// ControllerStoreKey is the store key used to store the controller.
	ControllerStoreKey = ControllerStoreKeyStruct{}
	// ReducedPrimariesStoreKey is the store key used to store the reduced
	// scope primary values.
	ReducedPrimariesStoreKey = ReducedPrimariesStoreStruct{}
	// PrimariesAlreadyChecked is the store key flag value used to notify that
	// the primaries for given process chain were already checked.
	PrimariesAlreadyChecked = primariesChecked{}
	// PreviousProcessStoreKey is the store key used to indicate what was the
	// previous query process.
	PreviousProcessStoreKey = PreviousProcess{}
	// AutoBeginStoreKey is the key that defines if query were started
	// with begin_transaction process.
	AutoBeginStoreKey = autoBegin{}
	// JustCreated is the store key that defines that the root scope's process is just after the creation process.
	JustCreated = created{}
)
