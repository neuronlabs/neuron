package query

import (
	"context"

	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal"
)

func beginTransactionFunc(ctx context.Context, s *Scope) error {
	if s.Err != nil {
		return nil
	}
	if s.tx != nil {
		return nil
	}

	transactioner, ok := s.repository().(Transactioner)
	if !ok {
		return nil
	}
	s.tx = Begin(ctx, s.c, nil)
	if err := s.tx.beginUniqueTransaction(transactioner, s.mStruct); err != nil {
		return err
	}

	s.StoreSet(internal.AutoBeginStoreKey, struct{}{})
	return nil
}

func commitOrRollbackFunc(ctx context.Context, s *Scope) error {
	if s.tx == nil {
		if log.Level().IsAllowed(log.LevelDebug3) {
			log.Debug3f("Scope[%s][%s] No transaction", s.ID(), s.Struct().Collection())
		}
		return nil
	}

	// check if the transaction was created by the process begin transaction.
	_, ok := s.StoreGet(internal.AutoBeginStoreKey)
	if !ok {
		if log.Level().IsAllowed(log.LevelDebug3) {
			log.Debug3f("SCOPE[%s][%s] No auto begin store key", s.ID(), s.Struct().Collection())
		}
		return nil
	}

	if s.Err == nil {
		if log.Level().IsAllowed(log.LevelDebug3) {
			log.Debug3f("SCOPE[%s][%s] Commit transaction[%s]", s.ID(), s.Struct().Collection(), s.tx.id)
		}
		return s.tx.Commit()
	}
	if log.Level().IsAllowed(log.LevelDebug3) {
		log.Debug3f("SCOPE[%s][%s] Rolling back transaction[%s]", s.ID(), s.Struct().Collection(), s.tx.id)
	}
	return s.tx.Rollback()
}
