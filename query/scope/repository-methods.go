package scope

// RepositoryMethoder is an interface that implements all possible repository methods interfaces
type RepositoryMethoder interface {
	Creater
	Getter
	Lister
	Patcher
	Deleter
}

// Creater is the repository interface that creates the value within the query.Scope
type Creater interface {
	Create(s *Scope) error
}

// Getter is the repository interface that Gets single query value
type Getter interface {
	Get(s *Scope) error
}

// Lister is the repository interface that Lists provided query values
type Lister interface {
	List(s *Scope) error
}

// Patcher is the repository interface that patches given query values
type Patcher interface {
	Patch(s *Scope) error
}

// Deleter is the interface for the repositories that deletes provided query value
type Deleter interface {
	Delete(s *Scope) error
}

/**

TRANSACTIONS (WIP)

*/

// Transactioner is the interface used for the transactions implementation
type Transactioner interface {
	Beginner
	Committer
	Rollbacker
}

// Beginner is the interface used for the distributed transaction to begin
type Beginner interface {
	Begin(s *Scope) error
}

// Committer is the interface used for committing the scope's transaction
type Committer interface {
	Commit(s *Scope) error
}

// Rollbacker is the interface used for rollbacks the interface transaction
type Rollbacker interface {
	Rollback(s *Scope) error
}
