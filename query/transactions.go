package query

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/repository"
)

// ErrRepositoryNotACommiter is an error returned when the repository doesn't implement Commiter interface
var (
	ErrRepositoryNotACommiter   = errors.New("Repository doesn't implement Commiter interface")
	ErrRepositoryNotARollbacker = errors.New("Repository doesn't implement Rollbacker interface")
)

// Tx is the scope's defined transaction
type Tx struct {
	ID      uuid.UUID `json:"id"`
	State   TxState   `json:"state"`
	Options TxOptions `json:"options"`
}

// TxOptions are the options for the Transaction
type TxOptions struct {
	Isolation IsolationLevel `json:"isolation"`
	ReadOnly  bool           `json:"read_only"`
}

// TxState defines the current transaction state
type TxState int

// Transaction state enums
const (
	_ TxState = iota
	TxBegin
	TxCommit
	TxRollback
	TxDone
)

func (s TxState) String() string {
	var str string
	switch s {
	case TxBegin:
		str = "begin"
	case TxCommit:
		str = "commit"
	case TxRollback:
		str = "rollback"
	default:
		str = "unknown"
	}
	return str
}

var _ fmt.Stringer = TxBegin

// MarshalJSON marshals the state into string value
// Implements the json.Marshaler interface
func (s *TxState) MarshalJSON() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalJSON unmarshals the state from the json string value
// Implements json.Unmarshaler interface
func (s *TxState) UnmarshalJSON(data []byte) error {
	str := string(data)
	switch str {
	case "commit":
		*s = TxCommit
	case "begin":
		*s = TxBegin
	case "rollback":
		*s = TxRollback
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

func (s *Scope) commitSingle(ctx context.Context, results chan<- interface{}) {
	var err error
	defer func() {
		if err != nil {
			results <- err
		} else {
			results <- struct{}{}
		}

		s.tx().State = TxCommit
	}()

	log.Debugf("Scope: %s, tx.State: %s", s.ID().String(), s.tx().State)
	if txn := s.tx(); txn.State != TxBegin {
		return
	}
	log.Debugf("Scope: %s for model: %s commiting tx: %s", s.ID().String(), s.Struct().Collection(), s.tx().ID.String())

	repo, err := repository.GetRepository(s.Controller(), s.Struct())
	if err != nil {
		log.Warningf("Repository not found for model: %v", s.Struct().Type().Name())
		err = ErrNoRepositoryFound
		return
	}

	cm, ok := repo.(Committer)
	if !ok {
		log.Debugf("Repository for model: '%s' doesn't implement Commiter interface", repo.RepositoryName())
		err = ErrRepositoryNotACommiter
		return
	}

	if err = cm.Commit(ctx, s); err != nil {
		return
	}
}

func (s *Scope) rollbackSingle(ctx context.Context, results chan<- interface{}) {
	var err error
	defer func() {
		s.tx().State = TxRollback
		if err != nil {
			results <- err

		} else {
			results <- struct{}{}
		}

	}()

	log.Debugf("Scope: %s, tx state: %s", s.ID().String(), s.tx().State)

	if txn := s.tx(); txn.State == TxRollback {
		return
	}

	log.Debugf("Scope: %s for model: %s rolling back tx: %s", s.ID().String(), s.Struct().Collection(), s.tx().ID.String())
	repo, err := repository.GetRepository(s.Controller(), s.Struct())
	if err != nil {
		log.Warningf("Repository not found for model: %v", s.Struct().Type().Name())
		err = ErrNoRepositoryFound
		return
	}

	rb, ok := repo.(Rollbacker)
	if !ok {
		log.Debugf("Repository for model: '%s' doesn't implement Rollbacker interface", repo.RepositoryName())
		err = ErrRepositoryNotARollbacker
		return
	}

	if err = rb.Rollback(ctx, s); err != nil {
		return
	}
}
