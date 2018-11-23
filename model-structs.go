package jsonapi

import (
	"fmt"
	"github.com/pkg/errors"
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

	// language is a field that contains the language information
	language *StructField

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

	// field that are ready for translations
	i18n []*StructField

	// ForeignKeys is a contianer for the foriegn keys for the relationships
	foreignKeys map[string]*StructField

	// filterKeys is a container for the filter keys
	filterKeys map[string]*StructField

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

// PrimaryField returns the primary index field for the model.
func (m *ModelStruct) PrimaryField() *StructField {
	return m.primary
}

// GetAttributeField - gets the attribute StructField for provided attribute.
func (m *ModelStruct) GetAttributeField(attr string) *StructField {
	return m.attributes[attr]
}

// Attribute gets the attribute on the base of the jsonapi tag name
func (m *ModelStruct) Attribute(attr string) (*StructField, bool) {
	sField, ok := m.attributes[attr]
	return sField, ok
}

// Attributes gets the attributes StructField for the given model.
func (m *ModelStruct) Attributes() (attrs []*StructField) {
	for _, attr := range m.attributes {
		attrs = append(attrs, attr)
	}
	return

}

// GetRelationshipField - gets the relationship StructField for provided relationship
func (m *ModelStruct) GetRelationshipField(relationship string) *StructField {
	return m.relationships[relationship]
}

func (m *ModelStruct) GetForeignKeyField(foreign string) *StructField {
	return m.foreignKeys[foreign]
}

// FilterKey gets the filter key struct field for the given model struct
// if the key is not found the returned boolean returns false
func (m *ModelStruct) FilterKey(key string) (*StructField, bool) {
	sField, ok := m.filterKeys[key]
	return sField, ok
}

// FilterKeys gets the filterkeys for the given model struct
func (m *ModelStruct) FilterKeys() map[string]*StructField {
	return m.filterKeys
}

func (m *ModelStruct) AllowClientID() bool {
	return m.primary.allowClientID()
}

// ListRelationshipFields - lists all relationship fields for given struct
func (m *ModelStruct) ListRelationshipFields() []*StructField {
	var fields []*StructField
	for _, field := range m.relationships {
		fields = append(fields, field)
	}
	return fields
}

// ListRelationshipNames - returns an array of the relationship names for given model
func (m *ModelStruct) ListRelationshipNames() (relationships []string) {
	for rel, fStruct := range m.relationships {
		if fStruct.isHidden() {
			continue
		}
		relationships = append(relationships, rel)
	}
	return
}

// GetFields - gets all structFields for given modelstruct (including primary).
func (m *ModelStruct) GetFields() []*StructField {
	return m.fields
}

// GetLanguageField - gets the Language field for given model.
// If the model does not support i18n the function returns nil structField.
func (m *ModelStruct) GetLanguageField() *StructField {
	return m.language
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

func (m *ModelStruct) FieldByName(name string) (field *StructField, ok bool) {
	for _, sfield := range m.fields {
		if sfield.refStruct.Name == name {
			return sfield, true
		}
	}
	return nil, false
}

func (m *ModelStruct) GetFieldByJSONAPIName(jsonapiName string) (*StructField, error) {
	for _, field := range m.fields {
		if field.jsonAPIName == jsonapiName {
			return field, nil
		}
	}
	return nil, fmt.Errorf("The model: '%s' does not contain jsonapi field: '%s'.", m.modelType.Name(), jsonapiName)
}

func (m *ModelStruct) SetModelURL(url string) error {
	return m.setModelURL(url)
}

func (m *ModelStruct) setBelongsToForeigns(v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Type() != m.modelType {
		return errors.Errorf("Invalid model type. Wanted: %v. Actual: %v", m.modelType.Name(), v.Type().Name())
	}
	for _, rel := range m.relationships {
		if rel.relationship != nil && rel.relationship.Kind == RelBelongsTo {
			relVal := v.FieldByIndex(rel.refStruct.Index)
			if reflect.DeepEqual(relVal.Interface(), reflect.Zero(relVal.Type()).Interface()) {
				continue
			}
			if relVal.Kind() == reflect.Ptr {
				relVal = relVal.Elem()
			}
			fkVal := v.FieldByIndex(rel.relationship.ForeignKey.refStruct.Index)
			relPrim := rel.relatedStruct.primary
			relPrimVal := relVal.FieldByIndex(relPrim.refStruct.Index)
			fkVal.Set(relPrimVal)
		}
	}
	return nil
}

func (m *ModelStruct) setBelongsToRelationWithFields(
	v reflect.Value,
	fields ...*StructField,
) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Type() != m.modelType {
		return errors.Errorf("Invalid model type. Wanted: %v. Actual: %v", m.modelType.Name(), v.Type().Name())
	}

	for _, field := range fields {

		rel, ok := m.relationships[field.jsonAPIName]
		if ok && rel.relationship != nil &&
			rel.relationship.Kind == RelBelongsTo {

			fkVal := v.FieldByIndex(rel.relationship.ForeignKey.refStruct.Index)
			relVal := v.FieldByIndex(rel.refStruct.Index)
			relType := relVal.Type()
			if relType.Kind() == reflect.Ptr {
				relType = relType.Elem()
			}

			if relVal.IsNil() {
				relVal.Set(reflect.New(relType))
			}
			relVal = relVal.Elem()

			relPrim := relVal.FieldByIndex(rel.relatedStruct.primary.refStruct.Index)
			relPrim.Set(fkVal)
		}
	}
	return nil
}

func (m *ModelStruct) setBelongsToForeignsWithFields(
	v reflect.Value, scope *Scope,
) ([]*StructField, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Type() != m.modelType {
		return nil, errors.Errorf("Invalid model type. Wanted: %v. Actual: %v", m.modelType.Name(), v.Type().Name())
	}
	fks := []*StructField{}
	for _, field := range scope.SelectedFields {
		rel, ok := m.relationships[field.jsonAPIName]
		if ok &&
			rel.relationship != nil &&
			rel.relationship.Kind == RelBelongsTo {
			relVal := v.FieldByIndex(rel.refStruct.Index)
			if reflect.DeepEqual(relVal.Interface(), reflect.Zero(relVal.Type()).Interface()) {
				continue
			}
			if relVal.Kind() == reflect.Ptr {
				relVal = relVal.Elem()
			}
			fkVal := v.FieldByIndex(rel.relationship.ForeignKey.refStruct.Index)
			relPrim := rel.relatedStruct.primary
			relPrimVal := relVal.FieldByIndex(relPrim.refStruct.Index)
			fkVal.Set(relPrimVal)
			fks = append(fks, rel.relationship.ForeignKey)
		}
	}
	return fks, nil
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

func (m *ModelStruct) getPrimaryValues(value reflect.Value) (primaries reflect.Value, err error) {
	primaryIndex := m.primary.getFieldIndex()
	switch value.Type().Kind() {
	case reflect.Slice:
		if value.Type().Elem().Kind() != reflect.Ptr {
			err = IErrUnexpectedType
			return
		}
		// create slice of values
		primaries = reflect.MakeSlice(reflect.SliceOf(m.primary.GetFieldType()), 0, value.Len())
		for i := 0; i < value.Len(); i++ {
			single := value.Index(i)
			if single.IsNil() {
				continue
			}
			single = single.Elem()
			primaryValue := single.Field(primaryIndex)
			if primaryValue.IsValid() {
				primaries = reflect.Append(primaries, primaryValue)
			}
		}
	case reflect.Ptr:
		primaryValue := value.Elem().Field(primaryIndex)
		if primaryValue.IsValid() {
			primaries = primaryValue
		} else {
			err = fmt.Errorf("Provided invalid Value for model: %v", m.GetType())
		}
	default:
		err = IErrUnexpectedType
	}
	return
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

// func (m *ModelStruct) initCheckStructFieldFlags() error {

// 	for _, field := range m.fields {

// 	}
// }

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
