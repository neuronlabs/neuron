package query

import (
	"context"

	"github.com/neuronlabs/neuron/common"

	"github.com/neuronlabs/neuron/internal"
)

var (
	// ProcessTransactionBegin is the process that begins the transaction.
	ProcessTransactionBegin = &Process{
		Name: "neuron:begin_transaction",
		Func: beginTransactionFunc,
	}

	// ProcessTransactionCommitOrRollback is the process that commits the scope's query
	// or rollbacks if the error occurred.
	ProcessTransactionCommitOrRollback = &Process{
		Name: "neuron:commit_or_rollback_transaction",
		Func: commitOrRollbackFunc,
	}
)

func beginTransactionFunc(ctx context.Context, s *Scope) error {
	_, ok := s.StoreGet(common.ProcessError)
	if ok {
		return nil
	}

	if tx := s.tx(); tx != nil {
		return nil
	}

	_, err := s.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	s.StoreSet(internal.AutoBeginStoreKey, struct{}{})

	return nil
}

func commitOrRollbackFunc(ctx context.Context, s *Scope) error {
	tx := s.Tx()
	if tx == nil {
		return nil
	}

	_, ok := s.StoreGet(internal.AutoBeginStoreKey)
	if !ok {
		return nil
	}

	_, ok = s.StoreGet(common.ProcessError)
	if !ok {
		if err := tx.CommitContext(ctx); err != nil {
			return err
		}
		return nil
	}

	err := tx.RollbackContext(ctx)
	if err != nil {
		return err
	}

	return nil
}
