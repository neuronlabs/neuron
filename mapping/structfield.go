package mapping

import (
	"github.com/neuronlabs/neuron/internal/models"
	"reflect"
)

// FieldKind is an enum that defines the following field type (i.e. 'primary', 'attribute')
type FieldKind models.FieldKind

const (
	// UnknownType is the undefined field kind
	UnknownType FieldKind = iota

	// KindPrimary is a 'primary' field
	KindPrimary

	// KindAttribute is an 'attribute' field
	KindAttribute

	// KindClientID is id set by client
	KindClientID

	// KindRelationshipSingle is a 'relationship' with single object
	KindRelationshipSingle

	// KindRelationshipMultiple is a 'relationship' with multiple objects
	KindRelationshipMultiple

	// KindForeignKey is the field type that is responsible for the relationships
	KindForeignKey

	// KindFilterKey is the field that is used only for special case filtering
	KindFilterKey

	// KindNested is the field's type that is signed as Nested
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
	return (*models.StructField)(s).ApiName()
}

// FieldIndex gets the field's index
func (s *StructField) FieldIndex() []int {
	return (*models.StructField)(s).FieldIndex()
}

// Nested returns the nested structure
func (s *StructField) Nested() *NestedStruct {
	nested := models.FieldsNested((*models.StructField)(s))
	if nested == nil {
		return nil
	}
	return (*NestedStruct)(nested)
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

// StoreSet sets into the store the value 'value' for given 'key'
func (s *StructField) StoreSet(key string, value interface{}) {
	(*models.StructField)(s).StoreSet(key, value)
}

// StoreGet gets the value from the store at the key: 'key'.
func (s *StructField) StoreGet(key string) (interface{}, bool) {
	return (*models.StructField)(s).StoreGet(key)
}

// StoreDelete deletes the store value at 'key'
func (s *StructField) StoreDelete(key string) {
	(*models.StructField)(s).StoreDelete(key)
}

// Name returns field's 'golang' name
func (s *StructField) Name() string {
	return (*models.StructField)(s).Name()
}

// FieldKind returns struct fields kind
func (s *StructField) FieldKind() FieldKind {
	return FieldKind((*models.StructField)(s).FieldKind())
}

// IsTimePointer checks if the field's type is a *time.time
func (s *StructField) IsTimePointer() bool {
	return models.FieldIsPtrTime((*models.StructField)(s))
}
