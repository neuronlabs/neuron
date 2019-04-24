package scope

// BeforeCreator is the interface used for hooks before the creation process
type BeforeCreator interface {
	HBeforeCreate(s *Scope) error
}

// AfterCreator is the interface that has a method used as a hook after the creation process
type AfterCreator interface {
	HAfterCreate(s *Scope) error
}

// BeforeGetter is the interface used as a hook before gettin value from api
type BeforeGetter interface {
	HBeforeGet(s *Scope) error
}

// BeforeLister is the interface used for before list hook
type BeforeLister interface {
	HBeforeList(s *Scope) error
}

// AfterGetter is the interface used as a hook after getting the value from api
type AfterGetter interface {
	HAfterGet(s *Scope) error
}

// AfterLister is the interface used as a after list hook
type AfterLister interface {
	HAfterList(s *Scope) error
}

// BeforePatcher is the interface used as a before patch hook
type BeforePatcher interface {
	HBeforePatch(s *Scope) error
}

// AfterPatcher is the interface used as a after patch hook
type AfterPatcher interface {
	HAfterPatch(s *Scope) error
}

// BeforeDeleter is the interface used as a before delete hook
type BeforeDeleter interface {
	HBeforeDelete(s *Scope) error
}

// AfterDeleter is the interface used as an after delete hook
type AfterDeleter interface {
	HAfterDelete(s *Scope) error
}
