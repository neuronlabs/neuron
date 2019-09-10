package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/controller"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"
	"github.com/neuronlabs/neuron-core/repository"

	"github.com/neuronlabs/neuron-core/internal"
)

// Tx is the scope's transaction model.
// It is bound to the single model type.
type Tx struct {
	ID      uuid.UUID `json:"id"`
	State   TxState   `json:"state"`
	Options TxOptions `json:"options"`

	root *Scope
}

// Commit commits the transaction.
func (t *Tx) Commit() error {
	return t.root.Commit()
}

// CommitContext commits the transaction with 'ctx' context.
func (t *Tx) CommitContext(ctx context.Context) error {
	return t.root.CommitContext(ctx)
}

// NewC creates new scope for given transaction.
func (t *Tx) NewC(c *controller.Controller, model interface{}) (*Scope, error) {
	return t.newC(context.Background(), c, model)
}

// NewContextC creates new scope for given transaction with the 'ctx' context.
func (t *Tx) NewContextC(ctx context.Context, c *controller.Controller, model interface{}) (*Scope, error) {
	return t.newC(ctx, c, model)
}

// NewContextModelC creates new scope for given 'model' structure and 'ctx' context. It also initializes scope's value.
// The value might be a slice of instances if 'isMany' is true or a single instance if false.
func (t *Tx) NewContextModelC(ctx context.Context, c *controller.Controller, model *mapping.ModelStruct, isMany bool) (*Scope, error) {
	return t.newModelC(ctx, c, model, isMany)
}

// NewModelC creates new scope for given model structure, and initializes it's value.
// The value might be a slice of instances if 'isMany' is true or a single instance if false.
func (t *Tx) NewModelC(c *controller.Controller, model *mapping.ModelStruct, isMany bool) (*Scope, error) {
	return t.newModelC(context.Background(), c, model, isMany)
}

// Rollback rolls back the transaction.
func (t *Tx) Rollback() error {
	return t.root.Rollback()
}

// RollbackContext rollsback the transaction with given 'ctx' context.
func (t *Tx) RollbackContext(ctx context.Context) error {
	return t.root.RollbackContext(ctx)
}

func (t *Tx) newC(ctx context.Context, c *controller.Controller, model interface{}) (*Scope, error) {
	s, err := newQueryScope(c, model)
	if err != nil {
		return nil, err
	}

	if err = t.setToScope(ctx, s); err != nil {
		return nil, err
	}

	return s, nil
}

func (t *Tx) newModelC(ctx context.Context, c *controller.Controller, model *mapping.ModelStruct, isMany bool) (*Scope, error) {
	s := newScopeWithModel(c, model, isMany)
	if err := t.setToScope(ctx, s); err != nil {
		return nil, err
	}

	return s, nil
}

func (t *Tx) setToScope(ctx context.Context, s *Scope) error {
	if s.Struct() != t.root.Struct() {
		// for models of different structure then the root one,
		// create new transactions with begin method and add it to the
		// subscope chain.
		repo, err := s.Controller().GetRepository(s.Struct())
		if err != nil {
			return err
		}

		tx, ok := t.root.transactions[repo]
		if ok {
			s.StoreSet(internal.TxStateStoreKey, tx)
			return nil
		}

		if _, err := s.begin(ctx, &t.Options, true); err != nil {
			return err
		}
		t.root.transactions[repo] = s.Tx()
		t.root.SubscopesChain = append(t.root.SubscopesChain, s)
	} else {
		// otherwise set the transaction to the store.
		s.StoreSet(internal.TxStateStoreKey, t)
	}
	return nil
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
		return errors.NewDet(class.QueryTxUnknownState, "unknown transaction state")
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
		return errors.NewDetf(class.QueryTxUnknownIsolationLevel, "unknown transaction isolation level: %s", string(data))
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

	if txn := s.tx(); txn.State != TxBegin {
		return
	}
	log.Debug3f("Scope[%s][%s] Commits transaction[%s]", s.ID().String(), s.Struct().Collection(), s.tx().ID.String())

	var repo repository.Repository
	repo, err = s.Controller().GetRepository(s.Struct())
	if err != nil {
		log.Warningf("Repository not found for model: %v", s.Struct().Type().Name())
		return
	}

	cm, ok := repo.(Transactioner)
	if !ok {
		log.Debugf("Repository for model: '%s' doesn't implement Commiter interface", repo.FactoryName())
		err = errors.NewDet(class.RepositoryNotImplementsTransactioner, "repository doesn't implement Transactioner")
		return
	}

	if err = cm.Commit(ctx, s); err != nil {
		return
	}
}

func (s *Scope) rollbackSingle(ctx context.Context, results chan<- interface{}) {
	var err error

	if txn := s.tx(); txn.State == TxRollback {
		results <- struct{}{}
		return
	}
	log.Debug3f("Scope[%s][%s] Rolls back transaction[%s]", s.ID().String(), s.Struct().Collection(), s.tx().ID.String())

	var repo repository.Repository
	repo, err = s.Controller().GetRepository(s.Struct())
	if err != nil {
		log.Warningf("Repository not found for model: %v", s.Struct().Type().Name())
		results <- err
		return
	}

	rb, ok := repo.(Transactioner)
	if !ok {
		log.Debugf("Repository for model: '%s' doesn't implement Rollbacker interface", repo.FactoryName())
		err = errors.NewDet(class.RepositoryNotImplementsTransactioner, "repository doesn't implement transactioner")
		results <- err
		return
	}

	if err = rb.Rollback(ctx, s); err != nil {
		results <- err
		return
	}
	s.tx().State = TxRollback
	results <- struct{}{}
}
