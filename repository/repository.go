package repository

import (
	"context"

	"github.com/neuronlabs/neuron/query"
)

// Repository is the interface used to execute the queries.
type Repository interface {
	// ID gets the repository unique identification.
	ID() string
	// Count counts the models for the provided query scope.
	Count(ctx context.Context, s *query.Scope) (int64, error)
	// Insert inserts models provided in the scope with provided field set.
	Insert(ctx context.Context, s *query.Scope) error
	// Find finds the models for provided query scope.
	Find(ctx context.Context, s *query.Scope) error
	// Update updates the query using a single model that  affects multiple database rows/entries.
	Update(ctx context.Context, s *query.Scope) (int64, error)
	// UpdateModels updates the models for provided query scope.
	UpdateModels(ctx context.Context, s *query.Scope) (int64, error)
	// Delete deletes the models defined by provided query scope.
	Delete(ctx context.Context, s *query.Scope) (int64, error)
}

// Exister is the interface used to check if given query object exists.
type Exister interface {
	Exists(context.Context, *query.Scope) (bool, error)
}

// Upserter is the repository interface that inserts or update on integration error given query values.
type Upserter interface {
	Upsert(ctx context.Context, s *query.Scope) error
}

// Transactioner is the interface used for the transactions.
type Transactioner interface {
	ID() string
	// Begin the scope's transaction.
	Begin(ctx context.Context, tx *query.Transaction) error
	// Commit the scope's transaction.
	Commit(ctx context.Context, tx *query.Transaction) error
	// Rollback the scope's transaction.
	Rollback(ctx context.Context, tx *query.Transaction) error
}

type Savepointer interface {
	Savepoint(ctx context.Context, tx *query.Transaction, name string) error
	RollbackSavepoint(ctx context.Context, tx *query.Transaction, name string) error
}
