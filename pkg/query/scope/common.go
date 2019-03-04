package scope

import (
	"github.com/pkg/errors"
)

var (
	ErrNilScopeValue     error = errors.New("Scope with nil value provided")
	ErrNoRepositoryFound error = errors.New("No repositories found for model.")
	ErrNoGetterRepoFound error = errors.New("No Getter repository possible for provided model")
	ErrNoListerRepoFound error = errors.New("No Lister repository possible for provided model")
	ErrNoPatcherFound    error = errors.New("The repository doesn't implement Patcher interface")
	ErrNoDeleterFound    error = errors.New("The repository doesn't implement Deleter interface")
)

// BeforeCreator is the interface used for hooks before the creation process
type wBeforeCreator interface {
	HBeforeCreate(s *Scope) error
}

// AfterCreatorR is the interface that has a method used as a hook after the creation process
type wAfterCreator interface {
	HAfterCreate(s *Scope) error
}

type wBeforeGetter interface {
	HBeforeGet(s *Scope) error
}

type wAfterGetter interface {
	HAfterGet(s *Scope) error
}

type wBeforeLister interface {
	HBeforeList(s *Scope) error
}

type wAfterLister interface {
	HAfterList(s *Scope) error
}

type wBeforePatcher interface {
	HBeforePatch(s *Scope) error
}

type wAfterPatcher interface {
	HAfterPatch(s *Scope) error
}

type wBeforeDeleter interface {
	HBeforeDelete(s *Scope) error
}

type wAfterDeleter interface {
	HAfterDeleter(s *Scope) error
}

// Repository is an interface that implements all possible
type repository interface {
	creater
	getlister
	patcher
	deleter
}

// creater is the repository interface that creates the value within the query.Scope
type creater interface {
	Create(scope *Scope) error
}

// Getlister is the repository that allows to get and list the query values
type getlister interface {
	getter
	lister
}

// getter is the repository interface that Gets single query value
type getter interface {
	Get(scope *Scope) error
}

// lister is the repository interface that Lists provided query values
type lister interface {
	List(scope *Scope) error
}

// patcher is the repository interface that patches given query values
type patcher interface {
	Patch(scope *Scope) error
}

// deleter is the interface for the repositories that deletes provided query value
type deleter interface {
	Delete(scope *Scope) error
}

// Transactions

type transactioner interface {
	beginner
	committer
	rollbacker
}

type beginner interface {
	Begin(s *Scope) error
}

type committer interface {
	Commit(s *Scope) error
}

type rollbacker interface {
	Rollback(s *Scope) error
}
