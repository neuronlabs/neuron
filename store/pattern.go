package store

// FindPattern is the pattern used for querying the store.
type FindPattern struct {
	// Suffix defines the search suffix.
	Suffix string
	// Prefix defines the search prefix.
	Prefix string
	// Limit limits the result maximum records number.
	Limit int
	// Offset shifts results with an integer offset.
	Offset int
}

// FindOption is a option func that changes find pattern.
type FindOption func(o *FindPattern)

// FindWithLimit sets the limit for the find pattern.
func FindWithLimit(limit int) FindOption {
	return func(o *FindPattern) {
		o.Limit = limit
	}
}

// FindWithOffset sets the offset for the find pattern.
func FindWithOffset(offset int) FindOption {
	return func(o *FindPattern) {
		o.Offset = offset
	}
}

// FindWithPrefix sets the prefix for the find pattern.
func FindWithPrefix(prefix string) FindOption {
	return func(o *FindPattern) {
		o.Prefix = prefix
	}
}

// FindWithSuffix sets the suffix for the find pattern.
func FindWithSuffix(suffix string) FindOption {
	return func(o *FindPattern) {
		o.Suffix = suffix
	}
}
