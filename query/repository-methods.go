package query

import (
	"context"

	"github.com/neuronlabs/neuron-core/mapping"
	"github.com/neuronlabs/neuron-core/repository"
)

// FullRepository is the interface that implements both repository CRUDRepository and the transactioner interfaces.
type FullRepository interface {
	CRUDRepository
	Transactioner
}

// CRUDRepository is an interface that implements all possible repository methods interfaces.
type CRUDRepository interface {
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
	repository.Repository
	// Begin the scope's transaction.
	Begin(ctx context.Context, tx *Tx, model *mapping.ModelStruct) error
	// Commit the scope's transaction.
	Commit(ctx context.Context, tx *Tx, model *mapping.ModelStruct) error
	// Rollback the scope's transaction.
	Rollback(ctx context.Context, tx *Tx, model *mapping.ModelStruct) error
}
