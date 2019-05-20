package scope

import (
	"context"
)

// BeforeCreator is the interface used for hooks before the creation process
type BeforeCreator interface {
	HBeforeCreate(ctx context.Context, s *Scope) error
}

// AfterCreator is the interface that has a method used as a hook after the creation process
type AfterCreator interface {
	HAfterCreate(ctx context.Context, s *Scope) error
}

// BeforeGetter is the interface used as a hook before gettin value from api
type BeforeGetter interface {
	HBeforeGet(ctx context.Context, s *Scope) error
}

// BeforeLister is the interface used for before list hook
type BeforeLister interface {
	HBeforeList(ctx context.Context, s *Scope) error
}

// AfterGetter is the interface used as a hook after getting the value from api
type AfterGetter interface {
	HAfterGet(ctx context.Context, s *Scope) error
}

// AfterLister is the interface used as a after list hook
type AfterLister interface {
	HAfterList(ctx context.Context, s *Scope) error
}

// BeforePatcher is the interface used as a before patch hook
type BeforePatcher interface {
	HBeforePatch(ctx context.Context, s *Scope) error
}

// AfterPatcher is the interface used as a after patch hook
type AfterPatcher interface {
	HAfterPatch(ctx context.Context, s *Scope) error
}

// BeforeDeleter is the interface used as a before delete hook
type BeforeDeleter interface {
	HBeforeDelete(ctx context.Context, s *Scope) error
}

// AfterDeleter is the interface used as an after delete hook
type AfterDeleter interface {
	HAfterDelete(ctx context.Context, s *Scope) error
}
