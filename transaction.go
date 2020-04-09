package neuron

import (
	"context"

	"github.com/neuronlabs/neuron-core/query"
)

// Begin starts new transaction with the default isolation.
func Begin() *query.Tx {
	return query.Begin()
}

// BeginCtx starts new transaction. Provided context 'ctx' is used until the transaction is committed or rolled back.
func BeginCtx(ctx context.Context, opts *query.TxOptions) *query.Tx {
	return query.BeginCtx(ctx, opts)
}

// TxFn is the function used to run WithTransaction.
type TxFn func(tx *query.Tx) error

// RunInTransaction creates a new transaction and handles rollback / commit based
// on the error return by the txFn.
func RunInTransaction(ctx context.Context, txOpts *query.TxOptions, txFn TxFn) (err error) {
	tx := query.BeginCtx(ctx, txOpts)
	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and repanic
			tx.Rollback()
			panic(p)
		} else if err != nil {
			// something went wrong, rollback
			tx.Rollback()
		} else {
			// all good, commit
			err = tx.Commit()
		}
	}()
	err = txFn(tx)
	return err
}
