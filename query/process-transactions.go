package query

import (
	"context"
	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal"
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
	_, ok := s.StoreGet(processErrorKey)
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
		if log.Level().IsAllowed(log.LDEBUG3) {
			log.Debug3f("Scope[%s][%s] No transaction", s.ID(), s.Struct().Collection())
		}
		return nil
	}

	_, ok := s.StoreGet(internal.AutoBeginStoreKey)
	if !ok {
		if log.Level().IsAllowed(log.LDEBUG3) {
			log.Debug3f("SCOPE[%s][%s] No autobegin store key", s.ID(), s.Struct().Collection())
		}
		return nil
	}

	_, ok = s.StoreGet(processErrorKey)
	if !ok {
		if log.Level().IsAllowed(log.LDEBUG3) {
			log.Debug3f("SCOPE[%s][%s] Commit transaction[%s]", s.ID(), s.Struct().Collection(), tx.ID)
		}
		if err := tx.CommitContext(ctx); err != nil {
			return err
		}
		return nil
	}
	if log.Level().IsAllowed(log.LDEBUG3) {
		log.Debug3f("SCOPE[%s][%s] Rolling back transaction[%s]", s.ID(), s.Struct().Collection(), tx.ID)
	}
	err := tx.RollbackContext(ctx)
	if err != nil {
		return err
	}

	return nil
}
