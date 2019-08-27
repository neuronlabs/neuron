package mapping

import (
	"reflect"

	"github.com/neuronlabs/neuron-core/internal/models"
)

// FieldKind is an enum that defines the following field type (i.e. 'primary', 'attribute').
type FieldKind models.FieldKind

const (
	// UnknownType is the undefined field kind.
	UnknownType FieldKind = iota

	// KindPrimary is a 'primary' field.
	KindPrimary

	// KindAttribute is an 'attribute' field.
	KindAttribute

	// KindClientID is id set by client.
	KindClientID

	// KindRelationshipSingle is a 'relationship' with single object.
	KindRelationshipSingle

	// KindRelationshipMultiple is a 'relationship' with multiple objects.
	KindRelationshipMultiple

	// KindForeignKey is the field type that is responsible for the relationships.
	KindForeignKey

	// KindFilterKey is the field that is used only for special case filtering.
	KindFilterKey

	// KindNested is the field's type that is signed as Nested.
	KindNested
)

// String implements fmt.Stringer interface.
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

// StructField represents a field structure with its json api parameters.
// and model relationships.
type StructField models.StructField

// FieldIndex gets the field's index.
func (s *StructField) FieldIndex() []int {
	return s.internal().FieldIndex()
}

// FieldKind returns struct fields kind.
func (s *StructField) FieldKind() FieldKind {
	return FieldKind(s.internal().FieldKind())
}

// IsTimePointer checks if the field's type is a *time.time.
func (s *StructField) IsTimePointer() bool {
	return s.internal().IsPtrTime()
}

// ModelStruct returns field's model struct.
func (s *StructField) ModelStruct() *ModelStruct {
	return (*ModelStruct)(s.internal().Struct())
}

// Name returns field's 'golang' name.
func (s *StructField) Name() string {
	return s.internal().Name()
}

// NeuronName returns the field's 'api' name.
func (s *StructField) NeuronName() string {
	return s.internal().NeuronName()
}

// Nested returns the nested structure.
func (s *StructField) Nested() *NestedStruct {
	nested := s.internal().Nested()
	if nested == nil {
		return nil
	}
	return (*NestedStruct)(nested)
}

// Relationship returns relationship for provided field.
func (s *StructField) Relationship() *Relationship {
	r := s.internal().Relationship()
	if r == nil {
		return nil
	}
	return (*Relationship)(r)
}

// ReflectField returns reflect.StructField related with this StructField.
func (s *StructField) ReflectField() reflect.StructField {
	return s.internal().ReflectField()
}

func (s *StructField) String() string {
	return s.NeuronName()
}

// StoreSet sets into the store the value 'value' for given 'key'.
func (s *StructField) StoreSet(key string, value interface{}) {
	s.internal().StoreSet(key, value)
}

// StoreGet gets the value from the store at the key: 'key'..
func (s *StructField) StoreGet(key string) (interface{}, bool) {
	return s.internal().StoreGet(key)
}

// StoreDelete deletes the store value at 'key'.
func (s *StructField) StoreDelete(key string) {
	s.internal().StoreDelete(key)
}

func (s *StructField) internal() *models.StructField {
	return (*models.StructField)(s)
}
