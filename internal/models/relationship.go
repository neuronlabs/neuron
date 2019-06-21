package models

import (
	"reflect"
)

// RelationshipKind is the enum used to define the Relationship's kind.
type RelationshipKind int

const (
	// RelUnknown is the unknown default relationship kind. States for the relationship internal errors.
	RelUnknown RelationshipKind = iota

	// RelBelongsTo is the enum value for the 'Belongs To' relationship.
	// This relationship kind states that the model containing the relationship field
	// contains also the foreign key of the related models.
	// The foreign key is a related model's primary field.
	RelBelongsTo

	// RelHasOne is the enum value for the 'Has One' relationship.
	// This relationship kind states that the model is in a one to one relationship with
	// the related model. It also states that the foreign key is located in the related model.
	RelHasOne

	// RelHasMany is the enum value for the 'Has Many' relationship.
	// This relationship kind states that the model is in a many to one relationship with the
	// related model. It also states that the foreign key is located in the related model.
	RelHasMany

	// RelMany2Many is the enum value for the 'Many To Many' relationship.
	// This relationship kind states that the model is in a many to many relationship with the
	// related model. This relationship requires the usage of the join model structure that contains
	// foreign keys of both related model types. The 'Relationship' struct foreign key should relate to the
	// model where the related field is stored - i.e. model 'user' has relationship field 'pets' to the model 'pet'
	// then the relationship pets foreign key should be a 'user id'. In order to get the foreign key of the related model
	// the relationship has also a field 'MtmForeignKey' which should be a 'pet id'.
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

// RelationshipOption defines the option on how to process the relationship.
type RelationshipOption int

// Relationship options
// TODO: prepare and define relationship options
const (
	Restrict RelationshipOption = iota
	NoAction
	Cascade
	SetNull
)

// Relationship is the structure that contains the relation's required field's
// kind, join model (if exists) and the process option (onDelete, onUpdate) as well
// as the definition for the related model's type 'mStruct'.
type Relationship struct {
	// kind is a relationship kind
	kind RelationshipKind

	// foreignKey represtents the foreign key field
	foreignKey *StructField

	// mtmRelatedForeignKey is the foreign key of the many2many related model.
	// The field is only used on the many2many relationships
	mtmRelatedForeignKey *StructField

	// joinModel is the join model used for the many2many relationships
	joinModel *ModelStruct

	// joinModelName is the collection name of the join model.
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

// ForeignKey returns relationships foreign key.
func (r *Relationship) ForeignKey() *StructField {
	return r.foreignKey
}

// IsManyToMany defines if the relaitonship is of many to many type.
func (r Relationship) IsManyToMany() bool {
	return r.isMany2Many()
}

// IsToMany defines if the relationship is of to many kind.
func (r Relationship) IsToMany() bool {
	return r.isToMany()
}

// IsToOne defines if the relationship is of to one type.
func (r Relationship) IsToOne() bool {
	return r.isToOne()
}

// JoinModel returns the join model for the given many2many relationship.
func (r *Relationship) JoinModel() *ModelStruct {
	return r.joinModel
}

// Kind returns relationship's kind.
func (r *Relationship) Kind() RelationshipKind {
	return r.kind
}

// ManyToManyForeignKey returns the foreign key of the many2many related model's.
func (r *Relationship) ManyToManyForeignKey() *StructField {
	return r.mtmRelatedForeignKey
}

// Struct returns relationship model *ModelStruct
func (r *Relationship) Struct() *ModelStruct {
	return r.mStruct
}

func (r Relationship) isToOne() bool {
	switch r.kind {
	case RelHasOne, RelBelongsTo:
		return true
	}
	return false
}

func (r Relationship) isToMany() bool {
	switch r.kind {
	case RelHasOne, RelBelongsTo, RelUnknown:
		return false
	}
	return true
}

func (r Relationship) isMany2Many() bool {
	return r.kind == RelMany2Many
}

// setForeignKey sets foreign key structfield.
func (r *Relationship) setForeignKey(s *StructField) {
	r.foreignKey = s
}

// SetKind sets the relationship kind
func (r *Relationship) setKind(kind RelationshipKind) {
	r.kind = kind
}
