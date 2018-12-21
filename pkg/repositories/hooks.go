package repositories

import (
	"github.com/kucjac/jsonapi/pkg/query"
)

// HookBeforeCreator is the Hook interface that allow to change the query before the creation
type HookBeforeCreator interface {
	HookBeforeCreate(scope *query.Scope) error
}

// HookAfterCreator is the Hook interface that allows to change the query after the creation
type HookAfterCreator interface {
	HookAfterCreate(scope *query.Scope) error
}

// HookBeforeReader is the Hook interface that allows to change the query before the read operation
type HookBeforeReader interface {
	HookBeforeRead(scope *query.Scope) error
}

// HookAfterReader is the Hook interface that allows to change the query after the read operation
type HookAfterReader interface {
	HookAfterRead(scope *query.Scope) error
}

// HookBeforePatcher is the Hook interface that allows to change the query before the patch
// operation
type HookBeforePatcher interface {
	HookBeforePatch(scope *query.Scope) error
}

// HookAfterPatcher is the Hook interface that allows to change the query after the patch
// operation
type HookAfterPatcher interface {
	HookAfterPatch(scope *query.Scope) error
}

// HookBeforeDeleter is the Hook interface that allows to change the query before the delete
// operation
type HookBeforeDeleter interface {
	HookBeforeDelete(scope *query.Scope) error
}

// HookAfterDeleter is the Hook interface that allows to change the query after the delete
// operation
type HookAfterDeleter interface {
	HookAfterDelete(scope *query.Scope) error
}
