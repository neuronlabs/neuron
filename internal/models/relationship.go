package models

import (
	"reflect"
)

// RelationshipKind is the enum of the Relationship kinds
type RelationshipKind int

// Relationship Kinds defined as the enums
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

// Relationship options
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

	// BackReference Fieldname is a field name that is back-reference
	// relationship in many2many relationships
	backReferenceForeignKeyName string

	// BackReferenceField is the backreferenced field in the many2many join model
	backReferenceForeignKey *StructField

	// joinModel is the join model used for the many2many relationships
	joinModel     *ModelStruct
	joinModelName string

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

// BackreferenceForeignKey returns the field that is backrefereneced in relationships models
func (r *Relationship) BackreferenceForeignKey() *StructField {
	return r.backReferenceForeignKey
}

// BackreferenceForeignKeyName returns relationships  backrefernce field name
func (r *Relationship) BackreferenceForeignKeyName() string {
	return r.backReferenceForeignKeyName
}

// ForeignKey returns relationships foreign key
func (r *Relationship) ForeignKey() *StructField {
	return r.foreignKey
}

// JoinModel returns the join model for the given many2many relationship
func (r *Relationship) JoinModel() *ModelStruct {
	return r.joinModel
}

// Kind returns relationships kind
func (r *Relationship) Kind() RelationshipKind {
	return r.kind
}

// SetBackreferenceForeignKey sets the Backreference Field for the relationship
func (r *Relationship) SetBackreferenceForeignKey(s *StructField) {
	r.backReferenceForeignKey = s
	r.backReferenceForeignKeyName = s.Name()
}

// SetForeignKey sets foreign key structfield
func (r *Relationship) SetForeignKey(s *StructField) {
	r.foreignKey = s
}

// SetKind sets the relationship kind
func (r *Relationship) SetKind(kind RelationshipKind) {
	r.kind = kind
}

// Struct returns relationship model *ModelStruct
func (r *Relationship) Struct() *ModelStruct {
	return r.mStruct
}

// RelationshipSetBackrefField sets the backreference field for the relationship
func RelationshipSetBackrefField(r *Relationship, backref *StructField) {
	r.backReferenceForeignKey = backref
}

// RelationshipMStruct returns relationship's modelstruc
func RelationshipMStruct(r *Relationship) *ModelStruct {
	return r.mStruct
}

// RelationshipGetKind returns relationship kind
func RelationshipGetKind(r *Relationship) RelationshipKind {
	return r.kind
}

// RelationshipForeignKey returns given relationship foreign key
func RelationshipForeignKey(r *Relationship) *StructField {
	return r.foreignKey
}

// IsToOne defines if the relationship is of to one type
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

// IsToMany defines if the relationship is of ToMany kind
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

// IsManyToMany defines if the relaitonship is of ManyToMany type
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
