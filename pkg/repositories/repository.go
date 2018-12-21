package repositories

import (
	"github.com/kucjac/jsonapi/pkg/query"
)

// Repository is an interface that implements all possible
type Repository interface {
	Creater
	GetLister
	Patcher
	Deleter
}

// Creater is the repository interface that creates the value within the query.Scope
type Creater interface {
	Create(scope *query.Scope) error
}

// GetLister is the repository that allows to get and list the query values
type GetLister interface {
	Getter
	Lister
}

// Getter is the repository interface that Gets single query value
type Getter interface {
	Get(scope *query.Scope) error
}

// Lister is the repository interface that Lists provided query values
type Lister interface {
	List(scope *query.Scope) error
}

// Patcher is the repository interface that patches given query values
type Patcher interface {
	Patch(scope *query.Scope) error
}

// Deleter is the interface for the repositories that deletes provided query value
type Deleter interface {
	Delete(scope *query.Scope) error
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
