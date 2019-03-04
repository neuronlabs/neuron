package mapping

import (
	"github.com/kucjac/jsonapi/pkg/internal/models"
	"reflect"
)

// FieldKind is an enum that defines the following field type (i.e. 'primary', 'attribute')
type FieldKind models.FieldKind

const (
	UnknownType FieldKind = iota
	// Primary is a 'primary' field
	KindPrimary

	// Attribute is an 'attribute' field
	KindAttribute

	// ClientID is id set by client
	KindClientID

	// RelationshipSingle is a 'relationship' with single object
	KindRelationshipSingle

	// RelationshipMultiple is a 'relationship' with multiple objects
	KindRelationshipMultiple

	// ForeignKey is the field type that is responsible for the relationships
	KindForeignKey

	// FilterKey is the field that is used only for special case filtering
	KindFilterKey

	KindNested
)

func (f FieldKind) String() string {
	switch f {
	case KindPrimary:
		return "Primary"
	case KindAttribute:
		return "Attribute"
	case KindClientID:
		return "ClientID"
	case KindRelationshipSingle, KindRelationshipMultiple:
		return "Relationship"
	case KindForeignKey:
		return "ForeignKey"
	case KindFilterKey:
		return "FilterKey"
	case KindNested:
		return "Nested"
	}

	return "Unknown"
}

// StructField represents a field structure with its json api parameters
// and model relationships.
type StructField models.StructField

// ApiName returns the field's 'api' name
func (s *StructField) ApiName() string {
	return (*models.StructField)(s).Name()
}

// FieldIndex gets the field's index
func (s *StructField) FieldIndex() []int {
	return (*models.StructField)(s).FieldIndex()
}

// Nested returns the nested structure
func (s *StructField) Nested() *NestedStruct {
	return (*NestedStruct)(models.FieldsNested((*models.StructField)(s)))
}

// Relationship returns relationship for provided field
func (s *StructField) Relationship() *Relationship {
	r := models.FieldRelationship((*models.StructField)(s))
	if r == nil {
		return nil
	}
	return (*Relationship)(r)
}

// ReflectField returns reflect.StructField related with this StructField
func (s *StructField) ReflectField() reflect.StructField {
	return (*models.StructField)(s).ReflectField()
}

// ModelStruct returns field's model struct
func (s *StructField) ModelStruct() *ModelStruct {
	return (*ModelStruct)(models.FieldsStruct((*models.StructField)(s)))
}

// Name returns field's 'golang' name
func (s *StructField) Name() string {
	return (*models.StructField)(s).Name()
}

// FieldKind returns struct fields kind
func (s *StructField) FieldKind() FieldKind {
	return FieldKind((*models.StructField)(s).FieldKind())
}
