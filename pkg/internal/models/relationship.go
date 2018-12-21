package models

import (
	"reflect"
)

// RelationshipKind is the enum of the Relationship kinds
type RelationshipKind int

const (
	RelUnknown RelationshipKind = iota
	RelBelongsTo
	RelHasOne
	RelHasMany
	RelMany2Many
)

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

// RelationshipOption defines the option on how to treat the relationship
type RelationshipOption int

const (
	Restrict RelationshipOption = iota
	NoAction
	Cascade
	SetNull
)

// By default the relationship many2many is local
// it can be synced with back-referenced relationship using the 'sync' flag
// The HasOne and HasMany is the relationship that by default is 'synced'
// The BelongsTo relationship is local by default.

// Relationship is the structure that uses
type Relationship struct {
	// kind is a relationship kind
	kind RelationshipKind

	// ForeignKey represtents the foreignkey field
	foreignKey *StructField

	// Sync is a flag that defines if the relationship opertaions
	// should be synced with the related many2many relationship
	// or the foreignkey in related foreign model
	sync *bool

	// BackReference Fieldname is a field name that is back-reference
	// relationship in many2many relationships
	backReferenceFieldname string

	// BackReferenceField
	backReferenceField *StructField

	// OnUpdate is a relationship option which determines
	// how the relationship should operate while updating the root object
	// By default it is set to Restrict
	onUpdate RelationshipOption

	// OnDelete is a relationship option which determines
	// how the relationship should operate while deleting the root object
	onDelete RelationshipOption

	// mStruct is the relationship model's structure
	mStruct *ModelStruct

	// modelType is the relationship's model Type
	modelType reflect.Type
}

// SetSync sets the relationship sync
func (r *Relationship) SetSync(b *bool) {
	r.sync = b
}

// SetKind sets the relationship kind
func (r *Relationship) SetKind(kind RelationshipKind) {
	r.kind = kind
}

// RelationshipSetBackrefField sets the backreference field for the relationship
func RelationshipSetBackrefField(r *Relationship, backref *StructField) {
	r.backReferenceField = backref
}

// RelationshipMStruct returns relationship's modelstruc
func RelationshipMStruct(r *Relationship) *ModelStruct {
	return r.mStruct
}

// RelationshipGetKing returns relationship kind
func RelationshipGetKind(r *Relationship) RelationshipKind {
	return r.kind
}

// RelationshipForeignKey returns given relationship foreign key
func RelationshipForeignKey(r *Relationship) *StructField {
	return r.foreignKey
}

// RelationshipSetBackrefFieldName sets the backreference fieldname
func RelationshipSetBackrefFieldName(r *Relationship, backrefName string) {
	r.backReferenceFieldname = backrefName
}

func (r Relationship) IsToOne() bool {
	return r.isToOne()
}

func (r Relationship) isToOne() bool {
	switch r.kind {
	case RelHasOne, RelBelongsTo:
		return true
	}
	return false
}

func (r Relationship) IsToMany() bool {
	return r.isToMany()
}

func (r Relationship) isToMany() bool {
	switch r.kind {
	case RelHasOne, RelBelongsTo, RelUnknown:
		return false
	}
	return true
}

func (r Relationship) IsManyToMany() bool {
	return r.isMany2Many()
}

func (r Relationship) isMany2Many() bool {
	switch r.kind {
	case RelMany2Many:
		return true
	}
	return false
}
