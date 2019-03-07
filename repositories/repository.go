package repositories

import (
	"github.com/kucjac/jsonapi/mapping"
	"github.com/kucjac/jsonapi/query/scope"
)

// RepositoryNamer is the interface used for the repositories to implement
// that defines it's name
type RepositoryNamer interface {
	RepositoryName() string
}

// Repository defines the repository
type Repository interface {
	New(model *mapping.ModelStruct) interface{}
	RepositoryNamer
	RepositoryMethoder
}

// Repository is an interface that implements all possible
type RepositoryMethoder interface {
	Creater
	GetLister
	Patcher
	Deleter
}

// Creater is the repository interface that creates the value within the query.Scope
type Creater interface {
	Create(scope *scope.Scope) error
}

// GetLister is the repository that allows to get and list the query values
type GetLister interface {
	Getter
	Lister
}

type Transactioner interface {
	Beginner
	Committer
	Rollbacker
}

type Beginner interface {
	Begin(s *scope.Scope) error
}

type Committer interface {
	Commit(s *scope.Scope) error
}

type Rollbacker interface {
	Rollback(s *scope.Scope) error
}

// Getter is the repository interface that Gets single query value
type Getter interface {
	Get(scope *scope.Scope) error
}

// Lister is the repository interface that Lists provided query values
type Lister interface {
	List(scope *scope.Scope) error
}

// Patcher is the repository interface that patches given query values
type Patcher interface {
	Patch(scope *scope.Scope) error
}

// Deleter is the interface for the repositories that deletes provided query value
type Deleter interface {
	Delete(scope *scope.Scope) error
}

// IsRepository is a quick check if the provided interface implements any of the repository
// interfaces
func IsRepository(r interface{}) bool {
	switch r.(type) {
	case Creater, GetLister, Patcher, Deleter:
		// if the provided 'r' repository implements any of the provided repository interfaces
		// Then it is treated as the repository
	default:
		return false
	}
	return true
}
