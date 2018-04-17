package jsonapi

import (
	"fmt"
	"reflect"
	"strings"
)

var maxNestedRelLevel int = 1

// JSONAPIType is an enum that defines the following field type (i.e. 'primary', 'attribute')
type JSONAPIType int

const (
	UnknownType JSONAPIType = iota
	// Primary is a 'primary' field
	Primary
	// Attribute is an 'attribute' field
	Attribute

	// RelationshipSingle is a 'relationship' with single object
	RelationshipSingle

	// RelationshipMultiple is a 'relationship' with multiple objects
	RelationshipMultiple
)

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

	// sortFieldCount is the number of sortable fields in the model
	sortFieldCount int

	// maximum included Count for this model struct
	thisIncludedCount   int
	nestedIncludedCount int
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

func (m *ModelStruct) getMaxIncludedCount() int {
	return m.thisIncludedCount + m.nestedIncludedCount
}

func (m *ModelStruct) getSortFieldCount() int {
	return m.sortFieldCount
}

func (m *ModelStruct) initComputeSortedFields() {
	for _, sField := range m.fields {
		if sField != nil && sField.canBeSorted() {
			m.sortFieldCount++
		}
	}
	return
}

func (m *ModelStruct) initComputeThisIncludedCount() {
	m.thisIncludedCount = len(m.relationships)
	return
}

func (m *ModelStruct) initComputeNestedIncludedCount(level int) int {
	var nestedCount int
	if level != 0 {
		nestedCount += m.thisIncludedCount
	}

	for _, relationship := range m.relationships {
		mStruct := cacheModelMap.Get(relationship.GetRelatedModelType())
		if mStruct == nil {
			panic(fmt.Sprintf("Model not mapped: %v", relationship.GetRelatedModelType()))
		}
		if level < maxNestedRelLevel {
			nestedCount += mStruct.initComputeNestedIncludedCount(level + 1)
		}
	}

	return nestedCount
}

func (m *ModelStruct) initCheckFieldTypes() error {
	for _, field := range m.fields {
		if field != nil {
			err := field.initCheckFieldType()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// StructField represents a field structure with its json api parameters
// and model relationships.
type StructField struct {
	// FieldName
	fieldName string

	// JSONAPIName is jsonapi field name - representation for json
	jsonAPIName string

	// fieldType
	jsonAPIType JSONAPIType

	// Given Field
	refStruct reflect.StructField

	// relatedModelType is a model type for the relationship
	relatedModelType reflect.Type

	// isListRelated
	isListRelated bool
}

// GetFieldIndex - gets the field index in the given model
func (s *StructField) GetFieldIndex() int {
	return s.getFieldIndex()
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
	return s.getRelatedModelType()
}

// GetJSONAPIType gets the JSONAPIType of the given struct field
func (s *StructField) GetJSONAPIType() JSONAPIType {
	return s.jsonAPIType
}

func (s *StructField) canBeSorted() bool {
	switch s.jsonAPIType {
	case RelationshipSingle, RelationshipMultiple, Attribute:
		return true
	}
	return false
}

func (s *StructField) getRelatedModelType() reflect.Type {
	return s.relatedModelType
}

func (s *StructField) getDereferencedType() reflect.Type {
	var t reflect.Type = s.refStruct.Type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func (s *StructField) getFieldIndex() int {
	return s.refStruct.Index[0]
}

func (s *StructField) initCheckFieldType() error {
	fieldType := s.refStruct.Type
	switch s.jsonAPIType {
	case Primary:
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		switch fieldType.Kind() {
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
			reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64:
		default:
			err := fmt.Errorf("Invalid primary field type: %s", fieldType)
			return err
		}
	case Attribute:
		// almost any type
		switch fieldType.Kind() {
		case reflect.Interface, reflect.Chan, reflect.Func, reflect.Invalid:
			err := fmt.Errorf("Invalid attribute field type: %v", fieldType)
			return err
		}

	case RelationshipSingle, RelationshipMultiple:
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		switch fieldType.Kind() {
		case reflect.Struct:
		case reflect.Slice:
			fieldType = fieldType.Elem()
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() != reflect.Struct {
				goto Gofallthrough
			}
		Gofallthrough:
			fallthrough
		default:
			err := fmt.Errorf("Invalid field type for the relationship.", fieldType)
			return err
		}
	}
	return nil
}
