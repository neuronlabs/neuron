package scope

// Kind is the enum defining the kind of scope
type Kind int

// Enums for the scope kind
const (
	RootKind Kind = iota
	IncludedKind
	RelationshipKind
	RelatedKind
	SubscopeKind
)

// Kind returns scope's kind
func (s *Scope) Kind() Kind {
	return s.kind
}

// SetKind sets the scope's kind
func (s *Scope) SetKind(kind Kind) {
	s.kind = kind
}

// IsSubscope checks if the given scope is a subscope
func (s *Scope) IsSubscope() bool {
	switch s.kind {
	case IncludedKind, SubscopeKind:
		return true
	default:
		return false
	}
}
