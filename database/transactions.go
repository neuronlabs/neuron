package database

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

// TxFunc is the execute function for the the RunInTransaction function.
type TxFunc func(db DB) error

// RunInTransaction runs specific function 'txFunc' within a transaction. If an error would return from that function the transaction would be rolled back.
// Otherwise it commits the changes. If an input 'db' is already within a transaction it would just execute txFunc for given transaction.
func RunInTransaction(ctx context.Context, db DB, options *query.TxOptions, txFunc TxFunc) error {
	if tx, ok := db.(*Tx); ok {
		return txFunc(tx)
	}

	// In all other cases create a new transaction and execute 'txFn'
	tx, err := Begin(ctx, db, options)
	if err != nil {
		return err
	}
	defer func() {
		p := recover()
		switch {
		case p != nil:
			// a panic occurred, rollback and repanic
			// A panic occurred, rollback and panic again.
			if er := tx.Rollback(); er != nil {
				log.Errorf("Rolling back on recover failed: %v", er)
			}
			panic(p)
		case err != nil:
			// If something went wrong, rollback
			if er := tx.Rollback(); er != nil {
				log.Errorf("Rolling back failed: %v", er)
			}
		default:
			// all good, commit
			// Everything is fine, commit given transaction.
			err = tx.Commit()
		}
	}()
	err = txFunc(tx)
	return err
}

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
			return nil, errors.WrapDetf(query.ErrTxDone, "transaction: '%s' is already done", tx.Transaction.ID.String())
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
		return errors.WrapDetf(query.ErrTxDone, "provided transaction: '%s' is already finished", t.Transaction.ID.String())
	}

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
		t.Transaction.State = query.TxFailed
		log.Debugf("Commit transaction: '%s' canceled by context: %v", t.Transaction.ID.String(), ctx.Err())
		return ctx.Err()
	case e := <-errChan:
		t.Transaction.State = query.TxFailed
		log.Debugf("Commit transaction: '%s' failed: %v", t.Transaction.ID.String(), e)
		return e
	case <-waitChan:
		t.Transaction.State = query.TxCommit
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
		return errors.WrapDetf(query.ErrTxDone, "provided transaction: '%s' is already finished", t.Transaction.ID)
	}

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
		t.Transaction.State = query.TxFailed
		log.Debugf("Rollback transaction: '%s' canceled by context: %v.", t.Transaction.ID.String(), ctx.Err())
		return ctx.Err()
	case e := <-errChan:
		t.Transaction.State = query.TxFailed
		log.Debugf("Rollback transaction: '%s' failed: %v.", t.Transaction.ID.String(), e)
		return e
	case <-waitChan:
		t.Transaction.State = query.TxRollback
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

var _ QueryFinder = &Tx{}

// QueryFind implements QueryFinder interface.
func (t *Tx) QueryFind(ctx context.Context, q *query.Scope) ([]mapping.Model, error) {
	if err := t.checkTransaction(); err != nil {
		return nil, err
	}
	if err := t.beginScopeTransaction(q); err != nil {
		return nil, err
	}
	models, err := queryFind(ctx, t, q)
	if err != nil {
		t.err = err
		return nil, err
	}
	return models, nil
}

var _ QueryGetter = &Tx{}

// QueryGet implements QueryGetter interface.
func (t *Tx) QueryGet(ctx context.Context, q *query.Scope) (mapping.Model, error) {
	if err := t.checkTransaction(); err != nil {
		return nil, err
	}
	if err := t.beginScopeTransaction(q); err != nil {
		return nil, err
	}
	model, err := queryGet(ctx, t, q)
	if err != nil {
		t.err = err
		return nil, err
	}
	return model, nil
}

// Insert implements DB interface.
func (t *Tx) Insert(ctx context.Context, mStruct *mapping.ModelStruct, models ...mapping.Model) error {
	if err := t.checkTransaction(); err != nil {
		return err
	}
	if len(models) == 0 {
		t.err = errors.Wrap(query.ErrNoModels, "nothing to insert")
		return t.err
	}
	if err := t.beginModelsTransaction(mStruct); err != nil {
		return err
	}
	s := query.NewScope(mStruct, models...)
	s.Transaction = t.Transaction
	if err := queryInsert(ctx, t, s); err != nil {
		t.err = err
		return err
	}
	return nil
}

var _ QueryInserter = &Tx{}

// InsertQuery implements QueryInserter interface.
func (t *Tx) InsertQuery(ctx context.Context, q *query.Scope) error {
	if err := t.checkTransaction(); err != nil {
		return err
	}
	if err := t.beginScopeTransaction(q); err != nil {
		return err
	}
	if err := queryInsert(ctx, t, q); err != nil {
		t.err = err
		return err
	}
	return nil
}

// Update implements DB interface.
func (t *Tx) Update(ctx context.Context, mStruct *mapping.ModelStruct, models ...mapping.Model) (int64, error) {
	if err := t.checkTransaction(); err != nil {
		return 0, err
	}
	if len(models) == 0 {
		t.err = errors.Wrap(query.ErrNoModels, "nothing to update")
		return 0, t.err
	}
	if err := t.beginModelsTransaction(mStruct); err != nil {
		return 0, err
	}
	s := query.NewScope(mStruct, models...)
	s.Transaction = t.Transaction
	affected, err := queryUpdate(ctx, t, s)
	if err != nil {
		t.err = err
		return 0, err
	}
	return affected, nil
}

var _ QueryUpdater = &Tx{}

// UpdateQuery implements QueryUpdater interface.
func (t *Tx) UpdateQuery(ctx context.Context, q *query.Scope) (int64, error) {
	if err := t.checkTransaction(); err != nil {
		return 0, err
	}
	if err := t.beginScopeTransaction(q); err != nil {
		t.err = err
		return 0, err
	}
	affected, err := queryUpdate(ctx, t, q)
	if err != nil {
		t.err = err
		return 0, err
	}
	return affected, nil
}

// Delete implements DB interface.
func (t *Tx) Delete(ctx context.Context, mStruct *mapping.ModelStruct, models ...mapping.Model) (int64, error) {
	if t.err != nil {
		return 0, t.err
	}
	if t.Transaction.State.Done() {
		t.err = errors.WrapDetf(query.ErrTxDone, "transaction: '%s' is already done", t.Transaction.ID.String())
		return 0, t.err
	}
	if len(models) == 0 {
		t.err = errors.Wrap(query.ErrNoModels, "nothing to delete")
		return 0, t.err
	}
	if err := t.beginModelsTransaction(mStruct); err != nil {
		return 0, err
	}
	s := query.NewScope(mStruct, models...)
	s.Transaction = t.Transaction
	affected, err := deleteQuery(ctx, t, s)
	if err != nil {
		t.err = err
		return 0, err
	}
	return affected, nil
}

var _ QueryDeleter = &Tx{}

// DeleteQuery implements QueryDeleter interface.
func (t *Tx) DeleteQuery(ctx context.Context, s *query.Scope) (int64, error) {
	if err := t.checkTransaction(); err != nil {
		return 0, err
	}
	if err := t.beginScopeTransaction(s); err != nil {
		return 0, err
	}
	affected, err := deleteQuery(ctx, t, s)
	if err != nil {
		t.err = err
		return 0, err
	}
	return affected, nil
}

// Refresh implements DB interface.
func (t *Tx) Refresh(ctx context.Context, mStruct *mapping.ModelStruct, models ...mapping.Model) error {
	if err := t.checkTransaction(); err != nil {
		return err
	}
	if len(models) == 0 {
		t.err = errors.WrapDetf(query.ErrNoModels, "nothing to refresh")
		return t.err
	}
	if err := t.beginModelsTransaction(mStruct); err != nil {
		return err
	}
	q := query.NewScope(mStruct, models...)
	q.Transaction = t.Transaction
	if err := refreshQuery(ctx, t, q); err != nil {
		t.err = err
		return err
	}
	return nil
}

var _ QueryRefresher = &Tx{}

// QueryRefresh implements QueryRefresher interface.
func (t *Tx) QueryRefresh(ctx context.Context, q *query.Scope) error {
	if err := t.checkTransaction(); err != nil {
		return err
	}
	if len(q.Models) == 0 {
		t.err = errors.WrapDetf(query.ErrNoModels, "nothing to refresh")
		return t.err
	}
	if err := t.beginScopeTransaction(q); err != nil {
		return err
	}
	if err := refreshQuery(ctx, t, q); err != nil {
		t.err = err
		return err
	}
	return nil
}

//
// Relations
//

// AddRelations implements DB interface.
func (t *Tx) AddRelations(ctx context.Context, model mapping.Model, relationField *mapping.StructField, relations ...mapping.Model) error {
	if err := t.checkTransaction(); err != nil {
		return err
	}
	mStruct, err := t.c.ModelStruct(model)
	if err != nil {
		t.err = err
		return err
	}
	if err := t.beginModelsTransaction(mStruct); err != nil {
		return err
	}
	q := query.NewScope(mStruct, model)
	q.Transaction = t.Transaction
	if err = queryAddRelations(ctx, t, q, relationField, relations...); err != nil {
		t.err = err
		return err
	}
	return nil
}

var _ QueryRelationAdder = &Tx{}

// QueryAddRelations implements QueryRelationAdder interface.
func (t *Tx) QueryAddRelations(ctx context.Context, s *query.Scope, relationField *mapping.StructField, relationModels ...mapping.Model) error {
	if err := t.checkTransaction(); err != nil {
		return err
	}
	if err := t.beginScopeTransaction(s); err != nil {
		return err
	}
	if err := queryAddRelations(ctx, t, s, relationField, relationModels...); err != nil {
		t.err = err
		return err
	}
	return nil
}

// SetRelations clears all 'relationField' for the input models and set their values to the 'relations'.
// The relation's foreign key must be allowed to set to null.
func (t *Tx) SetRelations(ctx context.Context, model mapping.Model, relationField *mapping.StructField, relations ...mapping.Model) error {
	if err := t.checkTransaction(); err != nil {
		return err
	}
	mStruct, err := t.c.ModelStruct(model)
	if err != nil {
		t.err = err
		return err
	}
	if err := t.beginModelsTransaction(mStruct); err != nil {
		return err
	}
	q := query.NewScope(mStruct, model)
	q.Transaction = t.Transaction
	if err = querySetRelations(ctx, t, q, relationField, relations...); err != nil {
		t.err = err
		return err
	}
	return nil
}

var _ QueryRelationSetter = &Tx{}

// QuerySetRelations implements QueryRelationSetter interface.
func (t *Tx) QuerySetRelations(ctx context.Context, q *query.Scope, relationField *mapping.StructField, relations ...mapping.Model) error {
	if err := t.checkTransaction(); err != nil {
		return err
	}
	if err := t.beginScopeTransaction(q); err != nil {
		return err
	}
	if err := querySetRelations(ctx, t, q, relationField, relations...); err != nil {
		t.err = err
		return err
	}
	return nil
}

// ClearRelations clears all 'relationField' relations for given input models.
// The relation's foreign key must be allowed to set to null.
func (t *Tx) ClearRelations(ctx context.Context, model mapping.Model, relationField *mapping.StructField) (int64, error) {
	if t.err != nil {
		return 0, t.err
	}
	if t.Transaction.State.Done() {
		t.err = errors.WrapDetf(query.ErrTxDone, "transaction: '%s' is already done", t.Transaction.ID.String())
		return 0, t.err
	}
	mStruct, err := t.c.ModelStruct(model)
	if err != nil {
		t.err = err
		return 0, err
	}
	if err := t.beginModelsTransaction(mStruct); err != nil {
		return 0, err
	}
	q := query.NewScope(mStruct, model)
	q.Transaction = t.Transaction
	affected, err := queryClearRelations(ctx, t, q, relationField)
	if err != nil {
		t.err = err
		return 0, err
	}
	return affected, nil
}

var _ QueryRelationClearer = &Tx{}

// QueryClearRelations implements QueryRelationClearer interface.
func (t *Tx) QueryClearRelations(ctx context.Context, q *query.Scope, relationField *mapping.StructField) (int64, error) {
	if err := t.checkTransaction(); err != nil {
		return 0, err
	}
	if err := t.beginScopeTransaction(q); err != nil {
		return 0, err
	}
	affected, err := queryClearRelations(ctx, t, q, relationField)
	if err != nil {
		t.err = err
		return 0, err
	}
	return affected, nil
}

// IncludeRelations implements DB interface.
func (t *Tx) IncludeRelations(ctx context.Context, mStruct *mapping.ModelStruct, models []mapping.Model, relationField *mapping.StructField, relationFieldset ...*mapping.StructField) error {
	if t.err != nil {
		return t.err
	}
	if t.Transaction.State.Done() {
		t.err = errors.WrapDetf(query.ErrTxDone, "transaction: '%s' is already done", t.Transaction.ID.String())
		return t.err
	}
	if err := t.beginModelsTransaction(mStruct); err != nil {
		return err
	}
	if err := queryIncludeRelation(ctx, t, mStruct, models, relationField, relationFieldset...); err != nil {
		t.err = err
		return err
	}
	return nil
}

// GetRelations implements DB interface.
func (t *Tx) GetRelations(ctx context.Context, mStruct *mapping.ModelStruct, models []mapping.Model, relationField *mapping.StructField, relationFieldset ...*mapping.StructField) ([]mapping.Model, error) {
	if t.err != nil {
		return []mapping.Model{}, t.err
	}
	if t.Transaction.State.Done() {
		t.err = errors.WrapDetf(query.ErrTxDone, "transaction: '%s' is already done", t.Transaction.ID.String())
		return []mapping.Model{}, t.err
	}
	if err := t.beginModelsTransaction(mStruct); err != nil {
		return []mapping.Model{}, err
	}
	relationModels, err := queryGetRelations(ctx, t, mStruct, models, relationField, relationFieldset...)
	if err != nil {
		t.err = err
		return []mapping.Model{}, err
	}
	return relationModels, nil
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

func (t *Tx) checkTransaction() error {
	if t.err != nil {
		return t.err
	}
	if t.Transaction.State.Done() {
		t.err = errors.WrapDetf(query.ErrTxDone, "transaction: '%s' is already done", t.Transaction.ID.String())
		return t.err
	}
	return nil
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
		t.err = errors.WrapDetf(query.ErrTxDone, "transaction: '%s' is already done", t.Transaction.ID.String())
		return tb
	}
	// create new scope and add it to the txQuery.

	s := query.NewScope(model, models...)
	s.Transaction = t.Transaction
	tb.scope = s

	if err := t.beginScopeTransaction(s); err != nil {
		return tb
	}
	return tb
}

func (t *Tx) beginScopeTransaction(s *query.Scope) error {
	// Get the repository mapped to given model.
	repo := getRepository(t.c, s)
	// Check if given repository is a transactioner.
	transactioner, ok := repo.(repository.Transactioner)
	if ok {
		if err := t.beginUniqueTransaction(transactioner, s.ModelStruct); err != nil {
			t.err = err
			return err
		}
		if s.Transaction == nil {
			s.Transaction = t.Transaction
		}
	} else {
		log.Debugf("Repository for model: '%s' doesn't support transactions", s.ModelStruct.String())
	}
	return nil
}

func (t *Tx) beginModelsTransaction(mStruct *mapping.ModelStruct) error {
	// Get the repository mapped to given model.
	repo := getModelRepository(t.c, mStruct)
	// Check if given repository is a transactioner.
	transactioner, ok := repo.(repository.Transactioner)
	if ok {
		if err := t.beginUniqueTransaction(transactioner, mStruct); err != nil {
			t.err = err
			return err
		}
	} else {
		log.Debugf("Repository for model: '%s' doesn't support transactions", mStruct.String())
	}
	return nil
}

func (t *Tx) beginUniqueTransaction(transactioner repository.Transactioner, model *mapping.ModelStruct) error {
	singleRepository := len(t.uniqueTransactions) == 1 || len(t.uniqueTransactions) == 0
	repoPosition := -1

	id := transactioner.ID()
	for i := 0; i < len(t.uniqueTransactions); i++ {
		if t.uniqueTransactions[i].id == id {
			repoPosition = i
			break
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
