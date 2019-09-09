package query

import (
	"context"
)

// FullRepository is the interface that implements both repository methoder
// and the transactioner intefaces.
type FullRepository interface {
	RepositoryMethoder
	Transactioner
}

// RepositoryMethoder is an interface that implements all possible repository methods interfaces.
type RepositoryMethoder interface {
	Counter
	Creator
	Getter
	Lister
	Patcher
	Deleter
}

// Counter is the interface used for query repositories that allows to count the result number
// for the provided query.
type Counter interface {
	Count(ctx context.Context, s *Scope) (int64, error)
}

// Creator is the repository interface that creates the value within the query.Scope.
type Creator interface {
	Create(ctx context.Context, s *Scope) error
}

// Getter is the repository interface that Gets single query value.
type Getter interface {
	Get(ctx context.Context, s *Scope) error
}

// Lister is the repository interface that Lists provided query values.
type Lister interface {
	List(ctx context.Context, s *Scope) error
}

// Patcher is the repository interface that patches given query values.
type Patcher interface {
	Patch(ctx context.Context, s *Scope) error
}

// Deleter is the interface for the repositories that deletes provided query value.
type Deleter interface {
	Delete(ctx context.Context, s *Scope) error
}

/**

TRANSACTIONS

*/

// Transactioner is the interface used for the transactions.
type Transactioner interface {
	// Begin the scope's transaction.
	Begin(ctx context.Context, s *Scope) error
	// Commit the scope's transaction.
	Commit(ctx context.Context, s *Scope) error
	// Rollback the scope's transaction.
	Rollback(ctx context.Context, s *Scope) error
}
