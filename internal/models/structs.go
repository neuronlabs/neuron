package models

import (
	"fmt"
	"github.com/kucjac/jsonapi/config"
	aerrors "github.com/kucjac/jsonapi/errors"
	"github.com/kucjac/jsonapi/flags"
	"github.com/kucjac/jsonapi/internal"
	"github.com/kucjac/jsonapi/log"
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

	// schemaName is the schema name set for given model
	schemaName string

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

	// repositoryName defines the model's repository name
	repositoryName string

	isAfterLister  bool
	isBeforeLister bool

	cfg *config.ModelConfig

	store map[string]interface{}
}

// AllowClientID returns boolean if the client settable id is allowed
func (m *ModelStruct) AllowClientID() bool {
	return m.primary.allowClientID()
}

// Attribute returns the attribute field for given string
func (m *ModelStruct) Attribute(field string) (*StructField, bool) {
	return StructAttr(m, field)
}

// Fields returns model's fields
func (m *ModelStruct) Fields() []*StructField {
	return m.fields
}

// StoreSet sets into the store the value 'value' for given 'key'
func (s *ModelStruct) StoreSet(key string, value interface{}) {
	if s.store == nil {
		s.store = make(map[string]interface{})
	}
	s.store[key] = value
}

// StoreGet gets the value from the store at the key: 'key'.
func (s *ModelStruct) StoreGet(key string) (interface{}, bool) {
	if s.store == nil {
		s.store = make(map[string]interface{})
	}
	v, ok := s.store[key]
	return v, ok
}

// StructFields return all the StructFields used in the ModelStruct
func (m *ModelStruct) StructFields() (fields []*StructField) {
	// add primary
	fields = append(fields, m.PrimaryField())

	// add attributes
	for _, f := range m.attributes {
		fields = append(fields, f)
	}

	// add relationships
	for _, f := range m.relationships {
		fields = append(fields, f)
	}

	// add foreignkeys
	for _, f := range m.foreignKeys {
		fields = append(fields, f)
	}

	// add i18n fields
	for _, f := range m.i18n {
		fields = append(fields, f)
	}

	if m.language != nil {
		// add language field
		fields = append(fields, m.language)
	}

	// add filterKey fields
	for _, f := range m.filterKeys {
		fields = append(fields, f)
	}

	return
}

// Flags return model's flags
func (m *ModelStruct) Flags() *flags.Container {
	return m.flags
}

// IsAfterLister returns the boolean if the given model implements
// AfterListerR interface
func (m *ModelStruct) IsAfterLister() bool {
	return m.isAfterLister
}

// IsBeforeLister returns the boolean if the given model implements
// BeforeListerR interface
func (m *ModelStruct) IsBeforeLister() bool {
	return m.isBeforeLister
}

// ForeignKey return model's foreign key
func (m *ModelStruct) ForeignKey(fk string) (*StructField, bool) {
	return StructForeignKeyField(m, fk)
}

// FilterKey return model's fitler key
func (m *ModelStruct) FilterKey(fk string) (*StructField, bool) {
	return StructForeignKeyField(m, fk)
}

// NewValueSingle creates and returns new value for the given model type
func (m *ModelStruct) NewValueSingle() interface{} {
	return m.newReflectValueSingle().Interface()
}

func (m *ModelStruct) NewReflectValueSingle() reflect.Value {
	return m.newReflectValueSingle()
}

func (m *ModelStruct) newReflectValueSingle() reflect.Value {
	return reflect.New(m.Type())
}

// PrimaryField returns model's primary struct field
func (m *ModelStruct) PrimaryField() *StructField {
	return m.primary
}

// RelationshipField returns the StructField for given raw field
func (m *ModelStruct) RelationshipField(field string) (*StructField, bool) {
	return StructRelField(m, field)
}

// RelationshipFields return structfields that are matched as relatinoships
func (m *ModelStruct) RelatinoshipFields() (rels []*StructField) {
	for _, rel := range m.relationships {
		rels = append(rels, rel)
	}
	return rels
}

// SchemaName returns model's schema name
func (m *ModelStruct) SchemaName() string {
	return m.schemaName
}

// SetConfig sets the config for given ModelStruct
func (m *ModelStruct) SetConfig(cfg *config.ModelConfig) error {
	log.Debugf("Setting Config: %v for model: %s", cfg, m.Collection())
	m.cfg = cfg

	m.repositoryName = cfg.Repository
	return nil
}

// Config returns config for the model
func (m *ModelStruct) Config() *config.ModelConfig {
	return m.cfg
}

// SetRepositoryName sets the repositoryName
func (m *ModelStruct) SetRepositoryName(repo string) {
	m.repositoryName = repo
}

// RepositoryName returns the repository name for given model
func (m *ModelStruct) RepositoryName() string {
	return m.repositoryName
}

// NewModelStruct creates new model struct for given type
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
func (m ModelStruct) ID() int {
	return m.id
}

// Type returns model's reflect.Type
func (m ModelStruct) Type() reflect.Type {
	return m.modelType
}

// Collection returns model's collection type
func (m ModelStruct) Collection() string {
	return m.collectionType
}

// UseI18n returns the bool if the model struct uses i18n.
func (m ModelStruct) UseI18n() bool {
	return m.language != nil
}

// SetSchemaName sets the schema name for the given model
func (m *ModelStruct) SetSchemaName(schema string) {
	m.schemaName = schema
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
			primaryValue := single.FieldByIndex(primaryIndex)
			if primaryValue.IsValid() {
				primaries = reflect.Append(primaries, primaryValue)
			}
		}
	case reflect.Ptr:
		primaryValue := value.Elem().FieldByIndex(primaryIndex)
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

func initComputeNestedIncludedCount(m *ModelStruct, level, maxNestedRelLevel int) int {
	var nestedCount int
	if level != 0 {
		nestedCount += m.thisIncludedCount
	}

	for _, relationship := range m.relationships {
		if level < maxNestedRelLevel {
			nestedCount += initComputeNestedIncludedCount(relationship.relationship.mStruct, level+1, maxNestedRelLevel)
		}
	}

	return nestedCount
}

// func (m *ModelStruct) initCheckStructFieldFlags() error {

// 	for _, field := range m.fields {

// 	}
// }

func (m *ModelStruct) InitComputeNestedIncludedCount(limit int) {

	m.nestedIncludedCount = initComputeNestedIncludedCount(m, 0, limit)
}

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
