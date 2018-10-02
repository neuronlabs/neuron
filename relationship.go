package jsonapi

type RelationshipKind int

const (
	RelUnknown RelationshipKind = iota
	RelBelongsTo
	RelHasOne
	RelHasMany
	RelMany2Many
	RelMany2ManyDisjoint
)

type Relationship struct {
	*StructField
	// Kind is a relationship kind
	Kind RelationshipKind

	// ForeignKey represtents the foreignkey field
	ForeignKey *StructField
}
