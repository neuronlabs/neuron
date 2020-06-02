package repository

import (
	"context"

	"github.com/neuronlabs/neuron/query"
)

// Repository is the interface used to execute the queries.
type Repository interface {
	Count(ctx context.Context, s *query.Scope) (int64, error)
	Insert(ctx context.Context, s *query.Scope) error
	Find(ctx context.Context, s *query.Scope) error
	Update(ctx context.Context, s *query.Scope) (int64, error)
	Delete(ctx context.Context, s *query.Scope) (int64, error)
}

// Exister is the interface used to check if given query object exists.
type Exister interface {
	Exists(context.Context, *query.Scope) (bool, error)
}

// Upserter is the repository interface that upserts given query values.
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
