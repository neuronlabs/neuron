package mapping

import (
	"github.com/neuronlabs/neuron/internal/models"
)

// RelationshipKind is the relation field's relationship kind enum.
type RelationshipKind models.RelationshipKind

const (
	// RelUnknown unknown relationship kind.
	RelUnknown RelationshipKind = iota

	// RelBelongsTo 'belongs to' relationship kind.
	RelBelongsTo

	// RelHasOne 'has one' relationship kind.
	RelHasOne

	// RelHasMany 'has many' relationship kind.
	RelHasMany

	// RelMany2Many 'many 2 many' relationship kind.
	RelMany2Many
)

// String implements fmt.Stringer interface.
func (r RelationshipKind) String() string {
	switch r {
	case RelUnknown:
	case RelBelongsTo:
		return "BelongsTo"
	case RelHasOne:
		return "HasOne"
	case RelHasMany:
		return "HasMany"
	case RelMany2Many:
		return "Many2Many"
	}
	return "Unknown"
}

// RelationshipOption is the relationship field's option enum.
type RelationshipOption models.RelationshipOption

// RelationshipOptions are used to define the default behaviour of the model's deleting and patching processes.
const (
	// Restrict is a restrict relationship option.
	Restrict RelationshipOption = iota
	NoAction
	Cascade
	SetNull
)

// By default the relationship many2many is local
// it can be synced with back-referenced relationship using the 'sync' flag
// The HasOne and HasMany is the relationship that by default is 'synced'
// The BelongsTo relationship is local by default.

// Relationship is a structure that defines the relation field's relationship.
type Relationship models.Relationship

// ModelStruct returns relationships model struct - related model structure.
func (r *Relationship) ModelStruct() *ModelStruct {
	mStruct := r.internal().Struct()
	if mStruct == nil {
		return nil
	}
	return (*ModelStruct)(mStruct)
}

// JoinModel is the join model used for the many2many relationship.
func (r *Relationship) JoinModel() *ModelStruct {
	return (*ModelStruct)(r.internal().JoinModel())
}

// Kind returns relationship Kind.
func (r *Relationship) Kind() RelationshipKind {
	k := r.internal().Kind()
	return RelationshipKind(k)
}

// ForeignKey returns foreign key for given relationship.
func (r *Relationship) ForeignKey() *StructField {
	fk := r.internal().ForeignKey()
	if fk == nil {
		return nil
	}
	return (*StructField)(fk)
}

// ManyToManyForeignKey returns the foreign key of the many2many related model's.
func (r *Relationship) ManyToManyForeignKey() *StructField {
	fk := r.internal().ManyToManyForeignKey()
	if fk == nil {
		return nil
	}
	return (*StructField)(fk)
}

func (r *Relationship) internal() *models.Relationship {
	return (*models.Relationship)(r)
}
