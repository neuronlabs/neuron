package query

import (
	"context"
)

// FullRepository is the interface that implements both repository CRUDRepository and the transactioner interfaces.
type FullRepository interface {
	CRUDRepository
	Transactioner
}

// CRUDRepository is an interface that implements all possible repository methods interfaces.
type CRUDRepository interface {
	Counter
	Exister
	Inserter
	Finder
	Updater
	Deleter
}

// Counter is the interface used for query repositories that allows to count the result number
// for the provided query.
type Counter interface {
	Count(ctx context.Context, s *Scope) (int64, error)
}

// Exister is the interface used to check if given query object exists.
type Exister interface {
	Exists(context.Context, *Scope) (bool, error)
}

// Inserter is the repository interface that creates the value within the query.Scope.
type Inserter interface {
	Insert(ctx context.Context, s *Scope) error
}

// Finder is the repository interface that Lists provided query values.
type Finder interface {
	Find(ctx context.Context, s *Scope) error
}

// Updater is the repository interface that Update given query values.
type Updater interface {
	Update(ctx context.Context, s *Scope) (int64, error)
}

// Upserter is the repository interface that upserts given query values.
type Upserter interface {
	Upsert(ctx context.Context, s *Scope) error
}

// Deleter is the interface for the repositories that deletes provided query value.
type Deleter interface {
	Delete(ctx context.Context, s *Scope) (int64, error)
}

/**

TRANSACTIONS

*/

// Transactioner is the interface used for the transactions.
type Transactioner interface {
	ID() string
	// Begin the scope's transaction.
	Begin(ctx context.Context, tx *Tx) error
	// Commit the scope's transaction.
	Commit(ctx context.Context, tx *Tx) error
	// Rollback the scope's transaction.
	Rollback(ctx context.Context, tx *Tx) error
}
