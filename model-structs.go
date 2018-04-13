package jsonapi

import (
	"fmt"
	"reflect"
	"strings"
)

var maxNestedRelLevel int = 1

// ModelStruct is a computed representation of the jsonapi models.
// Contain information about the model like the collection type,
// distinction of the field types (primary, attributes, relationships).
type ModelStruct struct {
	// modelType contain a reflect.Type information about given model
	modelType reflect.Type

	// collectionType is jsonapi 'type' for given model
	collectionType string

	// Primary is a jsonapi primary field
	primary *StructField

	// ClientID field that contains Client-Generated ID
	clientID *StructField

	// Attributes contain attribute fields
	attributes map[string]*StructField

	// Relationships contain jsonapi relationship fields
	// used to check and get if relationship exists in the model.
	// Can be heplful for validating url queries
	// Mapped StructField contain detailed information about given relationship
	relationships map[string]*StructField

	// Fields is a container of all public fields in the given model.
	// The field's index is the same as in the original model - for private or
	// non-settable fields the index would be nil
	fields []*StructField
}

// GetType - gets the reflect.Type of the model that modelstruct is based on.
func (m *ModelStruct) GetType() reflect.Type {
	return m.modelType
}

// GetCollectionType - gets the collection name ('type') for jsonapi.
func (m *ModelStruct) GetCollectionType() string {
	return m.collectionType
}

// GetPrimaryField - gets the primary field structField.
func (m *ModelStruct) GetPrimaryField() *StructField {
	return m.primary
}

// GetAttributeField - gets the attribute StructField for provided attribute.
func (m *ModelStruct) GetAttributeField(attr string) *StructField {
	return m.attributes[attr]
}

// GetRelationshipField - gets the relationship StructField for provided relationship
func (m *ModelStruct) GetRelationshipField(relationship string) *StructField {
	return m.relationships[relationship]
}

// GetFields - gets all structFields for given modelstruct (including primary).
func (m *ModelStruct) GetFields() []*StructField {
	return m.fields
}

// FieldByIndex return the StructField for given ModelStruct by the field index
// if the index is out of the range of Fields or there is no struct field with given index
// ( private field ) the function returns error
func (m *ModelStruct) FieldByIndex(index int) (*StructField, error) {
	if index > len(m.fields)-1 {
		return nil, fmt.Errorf("Index out of range: %v for fields in model: %s", index, m.modelType)
	}

	sField := m.fields[index]
	if sField == nil {
		return nil, fmt.Errorf("No field with index: %v found for model: %s", index, m.modelType)
	}

	return sField, nil
}

/** TO DELETE
// // FieldIndexByName returns the field index by provided name.
// // If not found return -1.
// func (m *ModelStruct) fieldIndexByName(name string) int {
// 	for i := 0; i < m.modelType.NumField(); i++ {
// 		if m.modelType.Field(i).Name == name {
// 			return i
// 		}
// 	}
// 	return -1
// }

*/

// there should be some helper which makes nested checks
// but root function should allow only to check once

func (m *ModelStruct) checkRelationshipHelper(relationship string) (errs []*ErrorObject, err error) {
	structField, ok := m.relationships[relationship]
	if !ok {
		split := strings.Split(relationship, annotationNestedSeperator)
		if len(split) == 1 {
			errs = append(errs, errNoRelationship(m.collectionType, relationship))
			return
		} else if len(split) > maxNestedRelLevel+1 {
			errs = append(errs, ErrTooManyNestedRelationships(relationship))
		}

		sepRelation := split[0]
		structField, ok = m.relationships[sepRelation]
		if !ok {
			errs = append(errs, errNoRelationship(m.collectionType, sepRelation))
			return
		}

		relM := cacheModelMap.Get(structField.relatedModelType)

		if relM == nil {
			err = errNoModelMappedForRel(structField.relatedModelType, m.modelType, structField.fieldName)
			errs = append(errs, ErrInternalError.Copy())
			return
		}

		var errorObjects []*ErrorObject
		errorObjects, err = relM.
			checkRelationshipHelper(strings.Join(split[1:], annotationNestedSeperator))
		errs = append(errs, errorObjects...)
	}
	return
}

// CheckRelationship is search for the modelstruct's relationship
// named as 'relationship'.
// The function checks and recursively get nested relationships.
// // Return multiple errors
func (m *ModelStruct) checkRelationship(relationship string) (errs []*ErrorObject, err error) {
	return m.checkRelationshipHelper(relationship)
}

// CheckRelationships is a method that checks if multiple relationships exists in the given model.
// Returns array of errors
func (m *ModelStruct) checkRelationships(relationships ...string,
) (errs []*ErrorObject, err error) {
	var errorObjects []*ErrorObject
	for _, relationship := range relationships {
		errorObjects, err = m.checkRelationship(relationship)
		errs = append(errs, errorObjects...)
		if err != nil {
			return
		}
	}
	return
}

// CheckAttribute - checks if given model contains given attributes. The attributes
// are checked for jsonapi manner.
func (m *ModelStruct) checkAttribute(attr string) *ErrorObject {
	_, ok := m.attributes[attr]
	if !ok {
		err := ErrInvalidQueryParameter.Copy()
		err.Detail = fmt.Sprintf("Object: '%v' does not have attribute: '%v'", m.collectionType, attr)
		return err
	}
	return nil
}

// CheckAttributesMultiErr checks if provided attributes exists in provided model.
// The attributes are checked as 'attr' tags in model structure.
// Returns multiple errors if occurs.
func (m *ModelStruct) checkAttributes(attrs ...string) (errs []*ErrorObject) {
	for _, attr := range attrs {
		if err := m.checkAttribute(attr); err != nil {
			errs = append(errs, err)
		}
	}
	return
}

func (m *ModelStruct) checkFields(fields ...string) (errs []*ErrorObject) {
	for _, field := range fields {
		_, hasAttribute := m.attributes[field]
		if hasAttribute {
			continue
		}
		_, hasRelationship := m.relationships[field]
		if !hasRelationship {
			errObject := ErrInvalidQueryParameter.Copy()
			errObject.Detail = fmt.Sprintf("Object: '%v', does not have field: '%v'.", m.collectionType, field)
			errs = append(errs, errObject)
		}
	}
	return
}

// StructField represents a field structure with its json api parameters
// and model relationships.
type StructField struct {
	// modelIndex defines field's index in the model used with reflect.Type.Field(index)
	modelIndex int

	// FieldName
	fieldName string

	// JSONAPIResKind is jsonapi specific resource type (i.e. primary, attr, relationship)
	jsonAPIResKind string

	// JSONAPIName is jsonapi field name - representation for json
	jsonAPIName string

	// IsPrimary is flag for primary keys
	isPrimary bool

	// Given Field
	refStruct reflect.StructField

	/** TO DELETE
	// if given structfield defines a relationship it should be non-nil
	// relationship *Relationship
	*/
	// relatedModelType is a model type for the relationship
	relatedModelType reflect.Type
}

// GetFieldIndex - gets the field index in the given model
func (s *StructField) GetFieldIndex() int {
	return s.modelIndex
}

// GetFieldName - gets the field name for given model
func (s *StructField) GetFieldName() string {
	return s.fieldName
}

// GetFieldType - gets the field's reflect.Type
func (s *StructField) GetFieldType() reflect.Type {
	return s.refStruct.Type
}

// GetRelatedModelType gets the reflect.Type of the related model
// used for relationship fields.
func (s *StructField) GetRelatedModelType() reflect.Type {
	return s.relatedModelType
}

/** TO DELETE
// // GetRelationship - gets the relationship for given structfield
// // The function returns also a check flag if the relationship was set for given structField
// func (s *StructField) GetRelationship() (Relationship, bool) {
// 	if s.relationship == nil {
// 		return Relationship{}, false
// 	}
// 	return *s.relationship, true
// }



// // RelationshipType defines an 'enum' for relationship type.
// type RelationshipType int

// const (
// 	// UnknownRelation - not defined relationship type
// 	UnknownRelation RelationshipType = iota

// 	// BelongsTo is a relationship of 'belongsto' type
// 	BelongsTo

// 	// HasOne is a relationship of one-to-one ('hasone') type
// 	HasOne

// 	// HasMany is a relationship of one-to-many ('hasone') type
// 	HasMany

// 	// ManyToMany is a relationship of many-to-many type
// 	ManyToMany
// )

// // Relationship contains information about the relationship between models.
// type Relationship struct {
// 	// Kind describes the relationship kind
// 	Type RelationshipType

// 	// RelatedType is a related model type.
// 	RelatedModelType reflect.Type
// }
*/
