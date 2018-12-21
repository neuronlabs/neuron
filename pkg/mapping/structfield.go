package mapping

import (
	"github.com/kucjac/jsonapi/pkg/internal/models"
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
type StructField struct {
	*models.StructField
}

// Nested returns the nested structure
func (s *StructField) Nested() *NestedStruct {
	return &NestedStruct{NestedStruct: models.FieldsNested(s.StructField)}
}

func (s *StructField) Relationship() *Relationship {
	r := models.FieldRelationship(s.StructField)
	if r == nil {
		return nil
	}
	return &Relationship{r}
}

// ModelStruct returns field's model struct
func (s *StructField) ModelStruct() *ModelStruct {
	return &ModelStruct{models.FieldsStruct(s.StructField)}
}

// FieldKind returns struct fields kind
func (s *StructField) FieldKind() FieldKind {
	return FieldKind(s.StructField.FieldKind())
}

// func (s *StructField) getRelationshipPrimariyValues(fieldValue reflect.Value,
// ) (primaries reflect.Value, err error) {
// 	if s.fieldType == RelationshipSingle || s.fieldType == RelationshipMultiple {
// 		return s.relatedStruct.getPrimaryValues(fieldValue)
// 	}
// 	// error
// 	err = fmt.Errorf("Provided field: '%v' is not a relationship: '%s'", s.fieldName, s.fieldType.String())
// 	return
// }
