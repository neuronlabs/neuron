package models

import (
	"fmt"
	aerrors "github.com/kucjac/jsonapi/errors"
	"github.com/kucjac/jsonapi/pkg/flags"
	"github.com/kucjac/jsonapi/pkg/internal"
	"github.com/pkg/errors"
	"reflect"
	"strings"
	"sync"
)

var ctr *counter = &counter{}

type counter struct {
	nextID int
	lock   sync.Mutex
}

func (c *counter) next() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.nextID += 1
	return c.nextID
}

// ModelStruct is a computed representation of the jsonapi models.
// Contain information about the model like the collection type,
// distinction of the field types (primary, attributes, relationships).
type ModelStruct struct {
	id int

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

	flags *flags.Container
}

func NewModelStruct(tp reflect.Type, collection string, flg *flags.Container) *ModelStruct {
	m := &ModelStruct{id: ctr.next(), modelType: tp, collectionType: collection, flags: flg}
	m.attributes = make(map[string]*StructField)
	m.relationships = make(map[string]*StructField)
	m.foreignKeys = make(map[string]*StructField)
	m.filterKeys = make(map[string]*StructField)
	m.collectionURLIndex = -1
	return m
}

// ID returns model structs index number
func (m *ModelStruct) ID() int {
	return m.id
}

// Type returns model's reflect.Type
func (m *ModelStruct) Type() reflect.Type {
	return m.modelType
}

// Collection returns model's collection type
func (m *ModelStruct) Collection() string {
	return m.collectionType
}

// UseI18n returns the bool if the model struct uses i18n.
func (m *ModelStruct) UseI18n() bool {
	return m.language != nil
}

// Flags return preset flags for the given model
func (m *ModelStruct) Flags() *flags.Container {
	return m.flags
}

// StructFieldByName returns field for provided name
// It matches both name and apiName
func StructFieldByName(m *ModelStruct, name string) *StructField {
	for _, field := range m.fields {
		if field.apiName == name || field.Name() == name {
			return field
		}
	}
	return nil
}

// StructAllFields return all model's fields
func StructAllFields(m *ModelStruct) []*StructField {
	return m.fields
}

// StructAppendField appends the field to the fieldset for the given model struct
func StructAppendField(m *ModelStruct, field *StructField) {
	m.fields = append(m.fields, field)
}

// StructSetAttr sets the attribute field for the provided model
func StructSetAttr(m *ModelStruct, attr *StructField) {
	m.attributes[attr.apiName] = attr
}

// StructSetRelField sets the relationship field for the model
func StructSetRelField(m *ModelStruct, relField *StructField) {
	m.relationships[relField.apiName] = relField
}

// StructSetForeignKey sets the foreign key for the model struct
func StructSetForeignKey(m *ModelStruct, fk *StructField) {
	m.foreignKeys[fk.apiName] = fk
}

// StructSetForeignKey sets the filter key for the model struct
func StructSetFilterKey(m *ModelStruct, fk *StructField) {
	m.filterKeys[fk.apiName] = fk
}

// StructSetPrimary
func StructSetPrimary(m *ModelStruct, primary *StructField) {
	m.primary = primary
}

// StructSetType
func StructSetType(m *ModelStruct, tp reflect.Type) {
	m.modelType = tp
}

// StructIsEqual checks if ModelStructs are equal
func StructIsEqual(first, second *ModelStruct) bool {
	return first.id == second.id
}

// StructSetUrl sets the ModelStruct's url value
func StructSetUrl(m *ModelStruct, url string) error {
	return m.setModelURL(url)
}

// StructPrimary gets the primary field for the given model
func StructPrimary(m *ModelStruct) *StructField {
	return m.primary
}

// StructI18n returns i18n struct field for given model.
func StructAppendI18n(m *ModelStruct, f *StructField) {
	m.i18n = append(m.i18n, f)
}

// StructSetLanguage sets the language field for given model
func StructSetLanguage(m *ModelStruct, f *StructField) {
	f.fieldFlags = f.fieldFlags | FLanguage
	m.language = f
}

func StructSetBelongsToForeigns(m *ModelStruct, v reflect.Value) error {
	return m.setBelongsToForeigns(v)
}

// StructAttr returns attribute for the provided structField
func StructAttr(m *ModelStruct, attr string) (*StructField, bool) {
	s, ok := m.attributes[attr]
	return s, ok
}

// StructRelField returns ModelsStruct relationship field if exists
func StructRelField(m *ModelStruct, relField string) (*StructField, bool) {
	s, ok := m.relationships[relField]
	return s, ok
}

// StructForeignKeyField returns ModelStruct foreign key field if exists
func StructForeignKeyField(m *ModelStruct, fk string) (*StructField, bool) {
	s, ok := m.foreignKeys[fk]
	return s, ok
}

// StructFilterKeyField returns ModelStruct filterKey
func StructFilterKeyField(m *ModelStruct, fk string) (*StructField, bool) {
	s, ok := m.filterKeys[fk]
	return s, ok
}

// StructSetBelongsToRelationWithFields sets the value to the Relationship value
func StructSetBelongsToRelationWithFields(m *ModelStruct, v reflect.Value, fields ...*StructField) error {
	return m.setBelongsToRelationWithFields(v, fields...)
}

func (m *ModelStruct) setBelongsToForeigns(v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Type() != m.modelType {
		return errors.Errorf("Invalid model type. Wanted: %v. Actual: %v", m.modelType.Name(), v.Type().Name())
	}
	for _, rel := range m.relationships {
		if rel.relationship != nil && rel.relationship.kind == RelBelongsTo {
			relVal := v.FieldByIndex(rel.reflectField.Index)
			if reflect.DeepEqual(relVal.Interface(), reflect.Zero(relVal.Type()).Interface()) {
				continue
			}
			if relVal.Kind() == reflect.Ptr {
				relVal = relVal.Elem()
			}
			fkVal := v.FieldByIndex(rel.relationship.foreignKey.reflectField.Index)
			relPrim := rel.relationship.mStruct.primary
			relPrimVal := relVal.FieldByIndex(relPrim.reflectField.Index)
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

		rel, ok := m.relationships[field.apiName]
		if ok && rel.relationship != nil &&
			rel.relationship.kind == RelBelongsTo {

			fkVal := v.FieldByIndex(rel.relationship.foreignKey.reflectField.Index)
			relVal := v.FieldByIndex(rel.reflectField.Index)
			relType := relVal.Type()
			if relType.Kind() == reflect.Ptr {
				relType = relType.Elem()
			}

			if relVal.IsNil() {
				relVal.Set(reflect.New(relType))
			}
			relVal = relVal.Elem()

			relPrim := relVal.FieldByIndex(rel.relationship.mStruct.primary.reflectField.Index)
			relPrim.Set(fkVal)
		}
	}
	return nil
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

// StructCollectionUrlIndex returns index for the collectionUrl
func StructCollectionUrlIndex(m *ModelStruct) int {
	return m.collectionURLIndex
}

// // CheckAttribute - checks if given model contains given attributes. The attributes
// // are checked for jsonapi manner.
// func (m *ModelStruct) checkAttribute(attr string) *apiErrors.ApiError {
// 	_, ok := m.Attributes[attr]
// 	if !ok {
// 		err := apiErrors.ErrInvalidQueryParameter.Copy()
// 		err.Detail = fmt.Sprintf("Object: '%v' does not have attribute: '%v'", m.collectionType, attr)
// 		return err
// 	}
// 	return nil
// }

// // CheckAttributesMultiErr checks if provided attributes exists in provided model.
// // The attributes are checked as 'attr' tags in model structure.
// // Returns multiple errors if occurs.
// func (m *ModelStruct) checkAttributes(attrs ...string) (errs []*apiErrors.ApiError) {
// 	for _, attr := range attrs {
// 		if err := m.checkAttribute(attr); err != nil {
// 			errs = append(errs, err)
// 		}
// 	}
// 	return
// }

func (m *ModelStruct) checkField(field string) (sField *StructField, err *aerrors.ApiError) {
	var hasAttribute, hasRelationship bool
	sField, hasAttribute = m.attributes[field]
	if hasAttribute {
		return sField, nil
	}
	sField, hasRelationship = m.relationships[field]
	if !hasRelationship {
		errObject := aerrors.ErrInvalidQueryParameter.Copy()
		errObject.Detail = fmt.Sprintf("Collection: '%v', does not have field: '%v'.", m.collectionType, field)
		return nil, errObject
	}
	return sField, nil
}

func (m *ModelStruct) checkFields(fields ...string) (errs []*aerrors.ApiError) {
	for _, field := range fields {
		_, err := m.checkField(field)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return
}

// StructCheckField is checks if the field exists within given modelstruct
func StructCheckField(m *ModelStruct, field string) (sField *StructField, err *aerrors.ApiError) {
	return m.checkField(field)
}

// StructSortScopeCount returns the count of the sort fieldsb
func StructSortScopeCount(m *ModelStruct) int {
	return m.sortScopeCount
}

// StructLanguage returns model's Language
func StructLanguage(m *ModelStruct) *StructField {
	return m.language
}

func StructMaxIncludedCount(m *ModelStruct) int {
	return m.thisIncludedCount + m.nestedIncludedCount
}

// PrimaryValues gets the primary values for the provided value
func (m *ModelStruct) PrimaryValues(value reflect.Value) (primaries reflect.Value, err error) {
	primaryIndex := m.primary.getFieldIndex()
	switch value.Type().Kind() {
	case reflect.Slice:
		if value.Type().Elem().Kind() != reflect.Ptr {
			err = internal.IErrUnexpectedType
			return
		}
		// create slice of values
		primaries = reflect.MakeSlice(reflect.SliceOf(m.primary.FieldType()), 0, value.Len())
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
			err = fmt.Errorf("Provided invalid Value for model: %v", m.Type())
		}
	default:
		err = internal.IErrUnexpectedType
	}
	return
}

func StructWorkingFieldCount(m *ModelStruct) int {
	return len(m.attributes) + len(m.relationships)
}

func InitComputeSortedFields(m *ModelStruct) {
	for _, sField := range m.fields {
		if sField != nil && sField.canBeSorted() {
			m.sortScopeCount++
		}
	}
	return
}

func InitComputeThisIncludedCount(m *ModelStruct) {
	m.thisIncludedCount = len(m.relationships)
	return
}

func InitComputeNestedIncludedCount(m *ModelStruct, level, maxNestedRelLevel int) int {
	var nestedCount int
	if level != 0 {
		nestedCount += m.thisIncludedCount
	}

	for _, relationship := range m.relationships {
		if level < maxNestedRelLevel {
			nestedCount += InitComputeNestedIncludedCount(relationship.relationship.mStruct, level+1, maxNestedRelLevel)
		}
	}

	return nestedCount
}

// func (m *ModelStruct) initCheckStructFieldFlags() error {

// 	for _, field := range m.fields {

// 	}
// }

func InitCheckFieldTypes(m *ModelStruct) error {
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
