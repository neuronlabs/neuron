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

	// ClientID is id set by client
	ClientID

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

	// sortScopeCount is the number of sortable fields in the model
	sortScopeCount int

	// maximum included Count for this model struct
	thisIncludedCount   int
	nestedIncludedCount int

	modelURL           string
	collectionURLIndex int
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
// func (m *ModelStruct) FieldByIndex(index int) (*StructField, error) {
// 	if index > len(m.fields)-1 {
// 		return nil, fmt.Errorf("Index out of range: %v for fields in model: %s", index, m.modelType)
// 	}

// 	sField := m.fields[index]
// 	if sField == nil {
// 		return nil, fmt.Errorf("No field with index: %v found for model: %s", index, m.modelType)
// 	}

// 	return sField, nil
// }

func (m *ModelStruct) SetModelURL(url string) error {
	return m.setModelURL(url)
}

func (m *ModelStruct) setModelURL(url string) error {
	splitted := strings.Split(url, "/")
	for i, v := range splitted {
		if v == m.collectionType {
			m.collectionURLIndex = i
			break
		}
	}
	if m.collectionURLIndex == -1 {
		err := fmt.Errorf("The url provided for model struct does not contain collection name. URL: '%s'. Collection: '%s'.", url, m.collectionType)
		return err
	}
	m.modelURL = url

	return nil
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

func (m *ModelStruct) checkField(field string) (sField *StructField, err *ErrorObject) {
	var hasAttribute, hasRelationship bool
	sField, hasAttribute = m.attributes[field]
	if hasAttribute {
		return sField, nil
	}
	sField, hasRelationship = m.relationships[field]
	if !hasRelationship {
		errObject := ErrInvalidQueryParameter.Copy()
		errObject.Detail = fmt.Sprintf("Collection: '%v', does not have field: '%v'.", m.collectionType, field)
		return nil, errObject
	}
	return sField, nil
}

func (m *ModelStruct) checkFields(fields ...string) (errs []*ErrorObject) {
	for _, field := range fields {
		_, err := m.checkField(field)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return
}

func (m *ModelStruct) getMaxIncludedCount() int {
	return m.thisIncludedCount + m.nestedIncludedCount
}

func (m *ModelStruct) getSortScopeCount() int {
	return m.sortScopeCount
}

func (m *ModelStruct) getWorkingFieldCount() int {
	return len(m.attributes) + len(m.relationships)
}

func (m *ModelStruct) initComputeSortedFields() {
	for _, sField := range m.fields {
		if sField != nil && sField.canBeSorted() {
			m.sortScopeCount++
		}
	}
	return
}

func (m *ModelStruct) initComputeThisIncludedCount() {
	m.thisIncludedCount = len(m.relationships)
	return
}

func (m *ModelStruct) initComputeNestedIncludedCount(level, maxNestedRelLevel int) int {
	var nestedCount int
	if level != 0 {
		nestedCount += m.thisIncludedCount
	}

	for _, relationship := range m.relationships {
		if level < maxNestedRelLevel {
			nestedCount += relationship.relatedStruct.initComputeNestedIncludedCount(level+1, maxNestedRelLevel)
		}
	}

	return nestedCount
}

func (m *ModelStruct) initCheckFieldTypes() error {
	err := m.primary.initCheckFieldType()
	if err != nil {
		return err
	}
	for _, field := range m.fields {
		if field != nil {
			err = field.initCheckFieldType()
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
	// model is the model struct that this field is part of.
	mStruct *ModelStruct

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
	relatedStruct    *ModelStruct

	// isListRelated
	isListRelated bool

	omitempty, iso8601, isTime, isPtrTime, noFilter bool
}

// GetFieldIndex - gets the field index in the given model
func (s *StructField) GetFieldIndex() int {
	return s.getFieldIndex()
}

// GetFieldName - gets the field name for given model
func (s *StructField) GetFieldName() string {
	return s.fieldName
}

// GetReflectStructField - gets the reflect.StructField for given field.
func (s *StructField) GetReflectStructField() reflect.StructField {
	return s.refStruct
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
			err := fmt.Errorf("Invalid primary field type: %s for the field: %s in model: %s.", fieldType, s.fieldName, s.mStruct.modelType.Name())
			return err
		}
	case Attribute:
		// almost any type
		switch fieldType.Kind() {
		case reflect.Interface, reflect.Chan, reflect.Func, reflect.Invalid:
			err := fmt.Errorf("Invalid attribute field type: %v for field: %s in model: %s", fieldType, s.fieldName, s.mStruct.modelType.Name())
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
				err := fmt.Errorf("Invalid slice type: %v, for the relationship: %v", fieldType,
					s.jsonAPIName)
				return err
			}
		default:
			err := fmt.Errorf("Invalid field type: %v, for the relationship.: %v", fieldType, s.jsonAPIName)
			return err
		}
	}
	return nil
}
