package store

// FindPattern is the pattern used for querying the store.
type FindPattern struct {
	Suffix string
	Prefix string
	Limit  int
	Offset int
}

// FindOption is a option func that changes find pattern.
type FindOption func(o *FindPattern)

// WithFindLimit sets the limit for the find pattern.
func WithFindLimit(limit int) FindOption {
	return func(o *FindPattern) {
		o.Limit = limit
	}
}

// WithFindOffset sets the offset for the find pattern.
func WithFindOffset(offset int) FindOption {
	return func(o *FindPattern) {
		o.Offset = offset
	}
}

// WithFindPrefix sets the prefix for the find pattern.
func WithFindPrefix(prefix string) FindOption {
	return func(o *FindPattern) {
		o.Prefix = prefix
	}
}

// WithFindSuffix sets the suffix for the find pattern.
func WithFindSuffix(suffix string) FindOption {
	return func(o *FindPattern) {
		o.Suffix = suffix
	}
}
