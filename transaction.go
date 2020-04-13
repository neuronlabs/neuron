package neuron

import (
	"context"

	"github.com/neuronlabs/neuron-core/controller"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/query"
)

// Begin starts new transaction with the default isolation.
func Begin() *query.Tx {
	return query.Begin(context.Background(), controller.Default(), nil)
}

// BeginCtx starts new transaction. Provided context 'ctx' is used until the transaction is committed or rolled back.
func BeginCtx(ctx context.Context, opts *query.TxOptions) *query.Tx {
	return query.Begin(ctx, controller.Default(), opts)
}

// TxFn is the function used to run WithTransaction.
type TxFn func(orm query.ORM) error

// RunInTransaction creates a new transaction and handles rollback / commit based
// on the error return by the txFn.
func RunInTransaction(ctx context.Context, orm query.ORM, txFn TxFn) (err error) {
	return runInTransaction(ctx, controller.Default(), orm, nil, txFn)
}

func runInTransaction(ctx context.Context, c *controller.Controller, orm query.ORM, txOpts *query.TxOptions, txFn TxFn) (err error) {
	switch db := orm.(type) {
	case *query.Tx:
		// Don't create new transaction if provided 'orm' is a transaction.
		return txFn(db)
	case *query.Scope:
		if db.Tx() != nil {
			// If given query scope is already inside of a transaction use it's transaction, so given expression
			// would be done within its transaction.
			return txFn(db.Tx())
		}
	}
	// In all other cases create a new transaction and execute 'txFn'
	tx := query.Begin(ctx, c, txOpts)
	defer func() {
		p := recover()
		switch {
		case p != nil:
			// A panic occurred, rollback and panic again.
			if er := tx.Rollback(); er != nil {
				log.Errorf("Rolling back on recover failed: %v", er)
			}
			panic(p)
		case err != nil:
			// If something went wrong, rollback.
			if er := tx.Rollback(); er != nil {
				log.Errorf("Rolling back failed: %v", er)
			}
		default:
			// Everything is fine, commit given transaction.
			err = tx.Commit()
		}
	}()
	err = txFn(tx)
	return err
}
