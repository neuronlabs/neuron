package repositories

import (
	"github.com/kucjac/jsonapi/pkg/query/scope"
)

// BeforeCreator is the interface used for hooks before the creation process
type BeforeCreator interface {
	WBeforeCreate(s *scope.Scope) error
}

// AfterCreator is the interface that has a method used as a hook after the creation process
type AfterCreator interface {
	WAfterCreate(s *scope.Scope) error
}

// BeforeGetter is the interface used as a hook before gettin value from api
type BeforeGetter interface {
	WBeforeGet(s *scope.Scope) error
}

// AfterGetterR is the interface used as a hook after getting the value from api
type AfterGetter interface {
	WAfterGetter(s *scope.Scope) error
}

type AfterLister interface {
	WAfterList(s *scope.Scope) error
}

type BeforePatcher interface {
	WBeforePatch(s *scope.Scope) error
}

type AfterPatcher interface {
	WAfterPatch(s *scope.Scope) error
}

type BeforeDeleter interface {
	WBeforeDelete(s *scope.Scope) error
}

type AfterDeleter interface {
	WAfterDelete(s *scope.Scope) error
}
