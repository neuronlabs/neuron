package query

import (
	"context"
)

/**

Creator

*/

// BeforeCreator is the interface used for hooks before the creation process.
type BeforeCreator interface {
	BeforeCreate(context.Context, *Scope) error
}

// AfterCreator is the interface that has a method used as a hook after the creation process.
type AfterCreator interface {
	AfterCreate(context.Context, *Scope) error
}

/**

Getter

*/

// BeforeGetter is the interface used as a hook before getting value.
type BeforeGetter interface {
	BeforeGet(context.Context, *Scope) error
}

// AfterGetter is the interface used as a hook after getting the value.
type AfterGetter interface {
	AfterGet(context.Context, *Scope) error
}

/**

Lister

*/

// BeforeLister is the interface used for before list hook.
type BeforeLister interface {
	BeforeList(context.Context, *Scope) error
}

// AfterLister is the interface used as a after list hook.
type AfterLister interface {
	AfterList(context.Context, *Scope) error
}

/**

Patcher

*/

// BeforePatcher is the interface used as a before patch hook.
type BeforePatcher interface {
	BeforePatch(ctx context.Context, s *Scope) error
}

// AfterPatcher is the interface used as a after patch hook.
type AfterPatcher interface {
	AfterPatch(ctx context.Context, s *Scope) error
}

/**

Deleter

*/

// BeforeDeleter is the interface used as a before delete hook.
type BeforeDeleter interface {
	BeforeDelete(ctx context.Context, s *Scope) error
}

// AfterDeleter is the interface used as an after delete hook.
type AfterDeleter interface {
	AfterDelete(ctx context.Context, s *Scope) error
}

/**

Counter

*/

// BeforeCounter is the interface used for before count hook.
type BeforeCounter interface {
	BeforeCount(ctx context.Context, s *Scope) error
}

// AfterCounter is the interface used for before count hook.
type AfterCounter interface {
	AfterCount(ctx context.Context, s *Scope) error
}
