package db

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/repository"
)

// Compile time check for the DB interface implementations.
var _ DB = &Tx{}

// Tx is an in-progress transaction orm. A transaction must end with a call to Commit or Rollback.
// After a call to Commit or Rollback all operations on the transaction fail with an error of class
type Tx struct {
	Transaction *query.Transaction

	c                  *controller.Controller
	err                error
	uniqueTransactions []*uniqueTx
}

// Begin starts new transaction with respect to the transaction context and transaction options with controller 'c'.
// If the 'db' is already a transaction the function
func Begin(ctx context.Context, db DB, options *query.TxOptions) (*Tx, error) {
	if tx, ok := db.(*Tx); ok {
		if tx.err != nil {
			return nil, tx.err
		}
		if tx.Transaction.State.Done() {
			return nil, errors.NewDetf(query.ClassTxDone, "transaction: '%s' is already done", tx.Transaction.ID.String())
		}
		return tx, nil
	}
	return begin(ctx, db.Controller(), options), nil
}

// Commit commits the transaction.
func (t *Tx) Commit() error {
	if len(t.uniqueTransactions) == 0 {
		log.Debugf("Commit transaction: %s, nothing to commit", t.Transaction.ID.String())
		return nil
	}
	if t.Transaction.State.Done() {
		return errors.NewDetf(query.ClassTxDone, "provided transaction: '%s' is already finished", t.Transaction.ID.String())
	}
	t.Transaction.State = query.TxCommit

	ctx, cancelFunc := context.WithCancel(t.Transaction.Ctx)
	defer cancelFunc()

	wg := &sync.WaitGroup{}
	txChan := t.produceUniqueTxChan(ctx, wg, t.getUniqueTransactions(false)...)

	errChan := make(chan error, 1)
	for tx := range txChan {
		go func(tx *uniqueTx) {
			defer wg.Done()
			log.Debug2f("Commit transaction '%s' for model: %s, %s", t.Transaction.ID, tx.model, tx.id)
			if err := tx.transactioner.Commit(ctx, t.Transaction); err != nil {
				errChan <- err
			}
		}(tx)
	}
	waitChan := make(chan struct{}, 1)
	go func() {
		log.Debug2f("Transaction: '%s' waiting for commits...", t.Transaction.ID)
		wg.Wait()
		waitChan <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		log.Debugf("Commit transaction: '%s' canceled by context: %v", t.Transaction.ID.String(), ctx.Err())
		return ctx.Err()
	case e := <-errChan:
		log.Debugf("Commit transaction: '%s' failed: %v", t.Transaction.ID.String(), e)
		return e
	case <-waitChan:
		log.Debugf("Commit transaction: '%s' with success", t.Transaction.ID.String())
	}
	return nil
}

// Rollback aborts the transaction.
func (t *Tx) Rollback() error {
	if len(t.uniqueTransactions) == 0 {
		log.Debugf("Rollback transaction: %s, nothing to rollback", t.Transaction.ID.String())
		return nil
	}
	if t.Transaction.State.Done() {
		return errors.NewDetf(query.ClassTxDone, "provided transaction: '%s' is already finished", t.Transaction.ID)
	}
	t.Transaction.State = query.TxRollback

	ctx, cancelFunc := context.WithCancel(t.Transaction.Ctx)
	defer cancelFunc()

	wg := &sync.WaitGroup{}
	txChan := t.produceUniqueTxChan(ctx, wg, t.getUniqueTransactions(true)...)

	errChan := make(chan error, 1)
	for tx := range txChan {
		go func(tx *uniqueTx) {
			defer wg.Done()
			log.Debug3f("Rollback transaction: '%s' for model: '%s'", t.Transaction.ID, tx.model)
			if err := tx.transactioner.Rollback(ctx, t.Transaction); err != nil {
				errChan <- err
			}
		}(tx)
	}

	waitChan := make(chan struct{}, 1)
	go func() {
		log.Debug3f("Transaction: '%s' waiting for rollbacks...", t.Transaction.ID)
		wg.Wait()
		waitChan <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		log.Debugf("Rollback transaction: '%s' canceled by context: %v.", t.Transaction.ID.String(), ctx.Err())
		return ctx.Err()
	case e := <-errChan:
		log.Debugf("Rollback transaction: '%s' failed: %v.", t.Transaction.ID.String(), e)
		return e
	case <-waitChan:
		log.Debugf("Rollback transaction: '%s' with success.", t.Transaction.ID.String())
	}
	return nil
}

// ID gets unique transaction uuid.
func (t *Tx) ID() uuid.UUID {
	return t.Transaction.ID
}

// Controller returns transaction controller.
func (t *Tx) Controller() *controller.Controller {
	return t.c
}

// Err returns current transaction runtime error.
func (t *Tx) Err() error {
	return t.err
}

// Options gets transaction TransactionOptions.
func (t *Tx) Options() query.TxOptions {
	return *t.Transaction.Options
}

// State gets current transaction Transaction.State.
func (t *Tx) State() query.TxState {
	return t.Transaction.State
}

// Query builds up a new query for given 'model'.
// The query is executed using transaction context.
func (t *Tx) Query(model *mapping.ModelStruct, models ...mapping.Model) Builder {
	return t.query(model, models...)
}

// QueryCtx builds up a new query for given 'model'.
// The query is executed using transaction context - provided context is used only for Builder purpose.
func (t *Tx) QueryCtx(ctx context.Context, model *mapping.ModelStruct, models ...mapping.Model) Builder {
	return t.query(model, models...)
}

func (t *Tx) Insert(ctx context.Context, models ...mapping.Model) error {
	if t.err != nil {
		return t.err
	}
	if t.Transaction.State.Done() {
		t.err = errors.NewDetf(query.ClassTxDone, "transaction: '%s' is already done", t.Transaction.ID.String())
		return t.err
	}

	if len(models) == 0 {
		return errors.New(query.ClassNoModels, "nothing to insert")
	}
	mStruct, err := t.c.ModelStruct(models[0])
	if err != nil {
		return err
	}
	s := query.NewScope(mStruct, models...)
	return Insert(ctx, t, s)
}

func (t *Tx) Update(ctx context.Context, models ...mapping.Model) (int64, error) {
	if t.err != nil {
		return 0, t.err
	}
	if t.Transaction.State.Done() {
		t.err = errors.NewDetf(query.ClassTxDone, "transaction: '%s' is already done", t.Transaction.ID.String())
		return 0, t.err
	}

	if len(models) == 0 {
		return 0, errors.New(query.ClassNoModels, "nothing to update")
	}
	mStruct, err := t.c.ModelStruct(models[0])
	if err != nil {
		return 0, err
	}
	s := query.NewScope(mStruct, models...)
	return Update(ctx, t, s)
}

func (t *Tx) Delete(ctx context.Context, models ...mapping.Model) (int64, error) {
	if t.err != nil {
		return 0, t.err
	}
	if t.Transaction.State.Done() {
		t.err = errors.NewDetf(query.ClassTxDone, "transaction: '%s' is already done", t.Transaction.ID.String())
		return 0, t.err
	}

	if len(models) == 0 {
		return 0, errors.New(query.ClassNoModels, "nothing to delete")
	}
	mStruct, err := t.c.ModelStruct(models[0])
	if err != nil {
		return 0, err
	}
	s := query.NewScope(mStruct, models...)
	return Delete(ctx, t, s)
}

func begin(ctx context.Context, c *controller.Controller, options *query.TxOptions) *Tx {
	if options == nil {
		options = &query.TxOptions{}
	}
	tx := &Tx{
		c: c,
		Transaction: &query.Transaction{
			ID:      uuid.New(),
			Ctx:     ctx,
			Options: options,
			State:   query.TxBegin,
		},
	}
	log.Debug2f("Begin transaction '%s'", tx.Transaction.ID.String())
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

func (t *Tx) query(model *mapping.ModelStruct, models ...mapping.Model) *txQuery {
	tb := &txQuery{tx: t}
	if t.err != nil {
		return tb
	}
	if t.Transaction.State.Done() {
		t.err = errors.NewDetf(query.ClassTxDone, "transaction: '%s' is already done", t.Transaction.ID.String())
		return tb
	}
	// create new scope and add it to the txQuery.

	s := query.NewScope(model, models...)
	s.Transaction = t.Transaction
	tb.scope = s

	// get the repository mapped to given model.
	repo := getRepository(t.c, s)
	// check if given repository is a transactioner.
	transactioner, ok := repo.(repository.Transactioner)
	if ok {
		// TODO: begin transaction just before executing processes.
		//		The relation processes shouldn't begin - i.e. AddRelationships - HasMany should begin related scope
		//		and not the root model's scope.
		if err := t.beginUniqueTransaction(transactioner, s.ModelStruct); err != nil {
			t.err = err
		}
	} else {
		log.Debug2f("Repository for model: '%s' doesn't support transactions", s.ModelStruct.String())
	}
	return tb
}

func (t *Tx) beginUniqueTransaction(transactioner repository.Transactioner, model *mapping.ModelStruct) error {
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
		log.Debug2f("Begin transaction '%s' for model: '%s'", t.Transaction.ID.String(), model.String())
		return transactioner.Begin(t.Transaction.Ctx, t.Transaction)
	}
	return nil
}

type uniqueTx struct {
	id            string
	transactioner repository.Transactioner
	model         *mapping.ModelStruct
}
