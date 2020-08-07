package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
)

// Transaction is the structure that defines the query transaction.
type Transaction struct {
	ID      uuid.UUID       `json:"id"`
	Ctx     context.Context `json:"context"`
	State   TxState         `json:"state"`
	Options *TxOptions      `json:"options"`
}

// TxOptions are the TransactionOptions for the Transaction
type TxOptions struct {
	Isolation IsolationLevel `json:"isolation"`
	ReadOnly  bool           `json:"read_only"`
}

// TxState defines the current transaction Transaction.State
type TxState int

// Done checks if current transaction is already finished.
func (t TxState) Done() bool {
	return t != TxBegin
}

// Transaction Transaction.State enums
const (
	TxBegin TxState = iota
	TxCommit
	TxRollback
	TxFailed
)

func (t TxState) String() string {
	var str string
	switch t {
	case TxBegin:
		str = "begin"
	case TxCommit:
		str = "commit"
	case TxRollback:
		str = "rollback"
	case TxFailed:
		str = "failed"
	default:
		str = "unknown"
	}
	return str
}

var _ fmt.Stringer = TxBegin

// MarshalJSON marshals the Transaction.State into string value
// Implements the json.Marshaler interface
func (t *TxState) MarshalJSON() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalJSON unmarshal the Transaction.State from the json string value
// Implements json.Unmarshaler interface
func (t *TxState) UnmarshalJSON(data []byte) error {
	str := string(data)
	switch str {
	case "commit":
		*t = TxCommit
	case "begin":
		*t = TxBegin
	case "rollback":
		*t = TxRollback
	case "unknown":
		*t = 0
	default:
		log.Errorf("Unknown transaction Transaction.State: %s", str)
		return errors.NewDet(ClassTxState, "unknown transaction Transaction.State")
	}
	return nil
}

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

// UnmarshalJSON unmarshal isolation level from the provided data
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
		return errors.NewDetf(ClassTxState, "unknown transaction isolation level: %s", string(data))
	}
	return nil
}

var _ fmt.Stringer = LevelDefault
