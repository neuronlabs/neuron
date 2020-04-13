package query

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/controller"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"
)

// Tx is an in-progress transaction. A transaction must end with a call to Commit or Rollback.
// After a call to Commit or Rollback all operations on the transaction fail with an error of class
type Tx struct {
	id uuid.UUID

	c     *controller.Controller
	ctx   context.Context
	state TxState

	options            *TxOptions
	err                error
	uniqueTransactions []*uniqueTx
}

// Begin startsC new transaction with respect to the 'ctx' context and transaction options 'options' and
// controller 'c'.
func Begin(ctx context.Context, c *controller.Controller, options *TxOptions) *Tx {
	return begin(ctx, c, options)
}

// ID gets unique transaction uuid.
func (t *Tx) ID() uuid.UUID {
	return t.id
}

// Controller returns transaction controller.
func (t *Tx) Controller() *controller.Controller {
	return t.c
}

// Err returns current transaction runtime error.
func (t *Tx) Err() error {
	return t.err
}

// Options gets transaction options.
func (t *Tx) Options() TxOptions {
	return *t.options
}

// State gets current transaction state.
func (t *Tx) State() TxState {
	return t.state
}

// Query builds up a new query for given 'model'.
// The query is executed using transaction context.
func (t *Tx) Query(model interface{}) Builder {
	return t.query(t.c, model)
}

// QueryCtx builds up a new query for given 'model'.
// The query is executed using transaction context - provided context is used only for Builder purpose.
func (t *Tx) QueryCtx(ctx context.Context, model interface{}) Builder {
	return t.query(t.c, model)
}

// Commit commits the transaction.
func (t *Tx) Commit() error {
	if len(t.uniqueTransactions) == 0 {
		log.Debugf("Commit transaction: %s, nothing to commit", t.id.String())
		return nil
	}
	if t.state.Done() {
		return errors.NewDetf(class.QueryTxDone, "provided transaction: '%s' is already finished", t.id.String())
	}
	t.state = TxCommit

	ctx, cancelFunc := context.WithCancel(t.ctx)
	defer cancelFunc()

	wg := &sync.WaitGroup{}
	txChan := t.produceUniqueTxChan(ctx, wg, t.getUniqueTransactions(false)...)

	errChan := make(chan error, 1)
	for tx := range txChan {
		go func(tx *uniqueTx) {
			defer wg.Done()
			log.Debug2f("Commit transaction '%s' for model: %s, %s", t.id, tx.model, tx.id)
			if err := tx.transactioner.Commit(ctx, t); err != nil {
				errChan <- err
			}
		}(tx)
	}
	waitChan := make(chan struct{}, 1)
	go func() {
		log.Debug2f("Transaction: '%s' waiting for commits...", t.id)
		wg.Wait()
		waitChan <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		log.Debugf("Commit transaction: '%s' canceled by context: %v", t.id.String(), ctx.Err())
		return ctx.Err()
	case e := <-errChan:
		log.Debugf("Commit transaction: '%s' failed: %v", t.id.String(), e)
		return e
	case <-waitChan:
		log.Debugf("Commit transaction: '%s' with success", t.id.String())
	}
	return nil
}

// Rollback aborts the transaction.
func (t *Tx) Rollback() error {
	if len(t.uniqueTransactions) == 0 {
		log.Debugf("Rollback transaction: %s, nothing to rollback", t.id.String())
		return nil
	}
	if t.state.Done() {
		return errors.NewDetf(class.QueryTxDone, "provided transaction: '%s' is already finished", t.id)
	}
	t.state = TxRollback

	ctx, cancelFunc := context.WithCancel(t.ctx)
	defer cancelFunc()

	wg := &sync.WaitGroup{}
	txChan := t.produceUniqueTxChan(ctx, wg, t.getUniqueTransactions(true)...)

	errChan := make(chan error, 1)
	for tx := range txChan {
		go func(tx *uniqueTx) {
			defer wg.Done()
			log.Debug3f("Rollback transaction: '%s' for model: '%s'", t.id, tx.model)
			if err := tx.transactioner.Rollback(ctx, t); err != nil {
				errChan <- err
			}
		}(tx)
	}

	waitChan := make(chan struct{}, 1)
	go func() {
		log.Debug3f("Transaction: '%s' waiting for rollbacks...", t.id)
		wg.Wait()
		waitChan <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		log.Debugf("Rollback transaction: '%s' canceled by context: %v.", t.id.String(), ctx.Err())
		return ctx.Err()
	case e := <-errChan:
		log.Debugf("Rollback transaction: '%s' failed: %v.", t.id.String(), e)
		return e
	case <-waitChan:
		log.Debugf("Rollback transaction: '%s' with success.", t.id.String())
	}
	return nil
}

func begin(ctx context.Context, c *controller.Controller, options *TxOptions) *Tx {
	if options == nil {
		options = &TxOptions{}
	}
	tx := &Tx{
		c:       c,
		id:      uuid.New(),
		ctx:     ctx,
		options: options,
		state:   TxBegin,
	}
	log.Debug2f("Begin transaction '%s'", tx.id.String())
	return tx
}

func (t *Tx) produceUniqueTxChan(ctx context.Context, wg *sync.WaitGroup, txs ...*uniqueTx) <-chan *uniqueTx {
	txChan := make(chan *uniqueTx, len(txs))
	go func() {
		defer close(txChan)
		for _, tx := range txs {
			func(tx *uniqueTx) {
				defer wg.Add(1)
				select {
				case <-ctx.Done():
					return
				default:
					txChan <- tx
				}
			}(tx)
		}
	}()
	return txChan
}

func (t *Tx) getUniqueTransactions(reverse bool) []*uniqueTx {
	if !reverse {
		return t.uniqueTransactions
	}
	txs := make([]*uniqueTx, len(t.uniqueTransactions))
	for i := 0; i < len(t.uniqueTransactions); i++ {
		j := len(t.uniqueTransactions) - 1 - i
		txs[i] = t.uniqueTransactions[j]
	}
	return txs
}

func (t *Tx) query(c *controller.Controller, model interface{}) *TxQuery {
	tb := &TxQuery{tx: t}
	if t.err != nil {
		return tb
	}
	if t.state.Done() {
		t.err = errors.NewDetf(class.QueryTxDone, "transaction: '%s' is already done", t.id.String())
		return tb
	}
	// create new scope and add it to the TxQuery.
	s, err := NewC(c, model)
	if err != nil {
		t.err = err
		return tb
	}
	tb.scope = s
	s.tx = t

	// get the repository mapped to given model.
	repo := s.repository()
	// check if given repository is a transactioner.
	transactioner, ok := repo.(Transactioner)
	if ok {
		if err = t.beginUniqueTransaction(transactioner, s.mStruct); err != nil {
			t.err = err
		}
	} else {
		log.Debug2f("Repository for model: '%s' doesn't support transactions", s.Struct().String())
	}
	return tb
}

func (t *Tx) beginUniqueTransaction(transactioner Transactioner, model *mapping.ModelStruct) error {
	singleRepository := true
	repoPosition := -1

	id := transactioner.ID()
	for i := 0; i < len(t.uniqueTransactions); i++ {
		if t.uniqueTransactions[i].id == id {
			repoPosition = i
		} else {
			singleRepository = false
		}
	}

	switch {
	case singleRepository && repoPosition != -1:
		// there is only one repository and it is 'transactioner'
		return nil
	case singleRepository:
		// there is only a single repository and it is not a 'transactioner'
		t.uniqueTransactions = append(t.uniqueTransactions, &uniqueTx{transactioner: transactioner, id: id, model: model})
	case repoPosition != -1:
		if len(t.uniqueTransactions)-1 == repoPosition {
			// the last repository in the transaction is of given type - do nothing
			return nil
		}
		// move transactioner from 'repoPosition' to the last
		t.uniqueTransactions = append(t.uniqueTransactions[:repoPosition], t.uniqueTransactions[repoPosition+1:]...)
		t.uniqueTransactions = append(t.uniqueTransactions, &uniqueTx{transactioner: transactioner, id: id, model: model})
	default:
		// current repository was not found - add it to the end of the queue.
		t.uniqueTransactions = append(t.uniqueTransactions, &uniqueTx{transactioner: transactioner, id: id, model: model})
	}

	if repoPosition == -1 {
		log.Debug2f("Begin transaction '%s' for model: '%s'", t.id.String(), model.String())
		return transactioner.Begin(t.ctx, t)
	}
	return nil
}

type uniqueTx struct {
	id            string
	transactioner Transactioner
	model         *mapping.ModelStruct
}

// TxOptions are the options for the Transaction
type TxOptions struct {
	Isolation IsolationLevel `json:"isolation"`
	ReadOnly  bool           `json:"read_only"`
}

// TxState defines the current transaction state
type TxState int

// Done checks if current transaction is already finished.
func (t TxState) Done() bool {
	return t == TxRollback || t == TxCommit
}

// Transaction state enums
const (
	TxBegin TxState = iota
	TxCommit
	TxRollback
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
	default:
		str = "unknown"
	}
	return str
}

var _ fmt.Stringer = TxBegin

// MarshalJSON marshals the state into string value
// Implements the json.Marshaler interface
func (t *TxState) MarshalJSON() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalJSON unmarshal the state from the json string value
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
		return errors.NewDetf(class.QueryTxUnknownIsolationLevel, "unknown transaction isolation level: %s", string(data))
	}
	return nil
}

var _ fmt.Stringer = LevelDefault
