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
	Creater
	Getter
	Lister
	Patcher
	Deleter
}

// Creater is the repository interface that creates the value within the query.Scope.
type Creater interface {
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
	Beginner
	Committer
	Rollbacker
}

// Beginner is the interface used for the distributed transaction to begin.
type Beginner interface {
	Begin(ctx context.Context, s *Scope) error
}

// Committer is the interface used for committing the scope's transaction.
type Committer interface {
	Commit(ctx context.Context, s *Scope) error
}

// Rollbacker is the interface used for rollbacks the transaction.
type Rollbacker interface {
	Rollback(ctx context.Context, s *Scope) error
}
