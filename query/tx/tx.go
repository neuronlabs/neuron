package tx

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/neuronlabs/neuron/log"
)

// Tx is the scope's defined transaction
type Tx struct {
	ID      uuid.UUID `json:"id"`
	State   State     `json:"state"`
	Options Options   `json:"options"`
}

// Options are the options for the Transaction
type Options struct {
	Isolation IsolationLevel `json:"isolation"`
	ReadOnly  bool           `json:"read_only"`
}

// State defines the current transaction state
type State int

// Transaction state enums
const (
	_ State = iota
	Begin
	Commit
	Rollback
	Done
)

func (s State) String() string {
	var str string
	switch s {
	case Begin:
		str = "begin"
	case Commit:
		str = "commit"
	case Rollback:
		str = "rollback"
	default:
		str = "unknown"
	}
	return str
}

var _ fmt.Stringer = Begin

// MarshalJSON marshals the state into string value
// Implements the json.Marshaler interface
func (s *State) MarshalJSON() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalJSON unmarshals the state from the json string value
// Implements json.Unmarshaler interface
func (s *State) UnmarshalJSON(data []byte) error {
	str := string(data)
	switch str {
	case "commit":
		*s = Commit
	case "begin":
		*s = Begin
	case "rollback":
		*s = Rollback
	case "unknown":
		*s = 0
	default:
		log.Errorf("Unknown transaction state: %s", str)
		return errors.New("Unknown transaction state")
	}
	return nil
}

var (
	// ErrTxAlreadyBegan notifies that the transaction had already began
	ErrTxAlreadyBegan = errors.New("Transaction already began")
)

// IsolationLevel is the
type IsolationLevel int

// Isolation level enums
const (
	LevelDefault IsolationLevel = iota
	LevelReadUncommitted
	LevelReadCommitted
	LevelWriteCommitted
	LevelRepeatableRead
	LevelSnapshot
	LevelSerializable
	LevelLinearizable
)

func (i IsolationLevel) String() string {
	switch i {
	case LevelDefault:
		return "default"
	case LevelReadUncommitted:
		return "read_uncommitted"
	case LevelReadCommitted:
		return "read_committed"
	case LevelWriteCommitted:
		return "write_committed"
	case LevelRepeatableRead:
		return "repeatable_read"
	case LevelSnapshot:
		return "snapshot"
	case LevelSerializable:
		return "serializable"
	case LevelLinearizable:
		return "linearizable"
	default:
		return "unknown"
	}
}

// MarshalJSON marshals the isolation level into json encoding
// Implements the json.Marshaller interface
func (i *IsolationLevel) MarshalJSON() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalJSON unmarshals isolation level from the provided data
// Implements json.Unmarshaler
func (i *IsolationLevel) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "default":
		*i = LevelDefault
	case "read_uncommitted":
		*i = LevelReadUncommitted
	case "read_committed":
		*i = LevelReadCommitted
	case "write_committed":
		*i = LevelWriteCommitted
	case "repeatable_read":
		*i = LevelRepeatableRead
	case "snapshot":
		*i = LevelSnapshot
	case "serializable":
		*i = LevelSerializable
	case "linearizable":
		*i = LevelLinearizable
	default:
		log.Debugf("Unknown transaction isolation level: %s", string(data))
		return errors.New("Unknown isolation level")
	}

	return nil

}

var _ fmt.Stringer = LevelDefault
