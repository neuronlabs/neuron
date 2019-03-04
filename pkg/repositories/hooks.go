package repositories

import (
	"github.com/kucjac/jsonapi/pkg/query/scope"
)

// BeforeCreator is the interface used for hooks before the creation process
type BeforeCreator interface {
	HBeforeCreate(s *scope.Scope) error
}

// AfterCreator is the interface that has a method used as a hook after the creation process
type AfterCreator interface {
	HAfterCreate(s *scope.Scope) error
}

// BeforeGetter is the interface used as a hook before gettin value from api
type BeforeGetter interface {
	HBeforeGet(s *scope.Scope) error
}

type BeforeLister interface {
	HBeforeList(s *scope.Scope) error
}

// AfterGetterR is the interface used as a hook after getting the value from api
type AfterGetter interface {
	HAfterGet(s *scope.Scope) error
}

type AfterLister interface {
	HAfterList(s *scope.Scope) error
}

type BeforePatcher interface {
	HBeforePatch(s *scope.Scope) error
}

type AfterPatcher interface {
	HAfterPatch(s *scope.Scope) error
}

type BeforeDeleter interface {
	HBeforeDelete(s *scope.Scope) error
}

type AfterDeleter interface {
	HAfterDelete(s *scope.Scope) error
}
