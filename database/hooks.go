package database

import (
	"context"
)

/**

Inserter

*/

// BeforeInserter is the interface used for hooks before the creation process.
type BeforeInserter interface {
	BeforeInsert(context.Context, DB) error
}

// AfterInserter is the interface that has a method used as a hook after the creation process.
type AfterInserter interface {
	AfterInsert(context.Context, DB) error
}

/**

Finder

*/

// AfterFinder is the interface used as a after find hook.
type AfterFinder interface {
	AfterFind(context.Context, DB) error
}

/**

Updater

*/

// BeforeUpdater is the interface used as a before patch hook.
type BeforeUpdater interface {
	BeforeUpdate(context.Context, DB) error
}

// AfterUpdater is the interface used as a after patch hook.
type AfterUpdater interface {
	AfterUpdate(context.Context, DB) error
}

/**

Deleter

*/

// BeforeDeleter is the interface used as a before delete hook.
type BeforeDeleter interface {
	BeforeDelete(context.Context, DB) error
}

// AfterDeleter is the interface used as an after delete hook.
type AfterDeleter interface {
	AfterDelete(context.Context, DB) error
}
