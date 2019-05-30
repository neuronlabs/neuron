package models

import (
	"fmt"
	"github.com/neuronlabs/neuron/config"
	aerrors "github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/flags"
	"github.com/neuronlabs/neuron/internal/namer"
	"github.com/neuronlabs/neuron/log"
	"github.com/pkg/errors"
	"reflect"
	"strings"
	"sync"
)

const (
	nestedIncludedCountKey = "neuron:nested_included_count"
	thisIncludedCountKey   = "neuron:this_included_count"
	namerFuncKey           = "neuron:namer_func"
)

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

	modelURL           string
	collectionURLIndex int

	flags *flags.Container

	isAfterLister  bool
	isBeforeLister bool

	cfg *config.ModelConfig

	store map[string]interface{}
}

// newModelStruct creates new model struct for given type
func newModelStruct(tp reflect.Type, collection string, flg *flags.Container) *ModelStruct {
	m := &ModelStruct{id: ctr.next(), modelType: tp, collectionType: collection, flags: flg}
	m.attributes = make(map[string]*StructField)
	m.relationships = make(map[string]*StructField)
	m.foreignKeys = make(map[string]*StructField)
	m.filterKeys = make(map[string]*StructField)
	m.collectionURLIndex = -1
	return m
}

// AllowClientID returns boolean if the client settable id is allowed
func (m *ModelStruct) AllowClientID() bool {
	return m.primary.allowClientID()
}

// Attribute returns the attribute field for given string
func (m *ModelStruct) Attribute(field string) (*StructField, bool) {
	return StructAttr(m, field)
}

// CheckField checks if the field exists within given modelstruct
func (m *ModelStruct) CheckField(field string) (sField *StructField, err *aerrors.ApiError) {
	return m.checkField(field)
}

// Collection returns model's collection type
func (m ModelStruct) Collection() string {
	return m.collectionType
}

// Config returns config for the model
func (m *ModelStruct) Config() *config.ModelConfig {
	return m.cfg
}

// Fields returns model's fields
func (m *ModelStruct) Fields() []*StructField {
	return m.fields
}

// Flags return model's flags
func (m *ModelStruct) Flags() *flags.Container {
	return m.flags
}

// ForeignKey return model's foreign key
func (m *ModelStruct) ForeignKey(fk string) (*StructField, bool) {
	return StructForeignKeyField(m, fk)
}

// ForeignKeys return ForeignKey Structfields for the given model
func (m *ModelStruct) ForeignKeys() (fks []*StructField) {
	for _, f := range m.foreignKeys {
		fks = append(fks, f)
	}
	return
}

// FieldCount gets the field count for given model
func (m *ModelStruct) FieldCount() int {
	return len(m.attributes) + len(m.relationships)
}

// FilterKey return model's fitler key
func (m *ModelStruct) FilterKey(fk string) (*StructField, bool) {
	return StructFilterKeyField(m, fk)
}

// ID returns model structs index number
func (m ModelStruct) ID() int {
	return m.id
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

// LanguageField returns model's Language
func (m *ModelStruct) LanguageField() *StructField {
	return m.language
}

// MaxIncludedCount gets the maximum included count prepared for given model
func (m *ModelStruct) MaxIncludedCount() int {
	thisIncludedInterface, ok := m.StoreGet(thisIncludedCountKey)
	if !ok {
		return 0
	}

	thisIncluded := thisIncludedInterface.(int)

	nestedIncludedInterface, ok := m.StoreGet(nestedIncludedCountKey)
	if !ok {
		return thisIncluded
	}

	return thisIncluded + nestedIncludedInterface.(int)
}

// NamerFunc returns namer func for the given Model
func (m *ModelStruct) NamerFunc() namer.Namer {
	v, _ := m.StoreGet(namerFuncKey)
	return v.(namer.Namer)

}

// NewValueSingle creates and returns new value for the given model type
func (m *ModelStruct) NewValueSingle() interface{} {
	return m.newReflectValueSingle().Interface()
}

// NewValueMany creates and returns a model's new slice of pointers to values
func (m *ModelStruct) NewValueMany() interface{} {
	return m.NewReflectValueMany().Interface()
}

// NewReflectValueSingle creates and returns a model's new single value
func (m *ModelStruct) NewReflectValueSingle() reflect.Value {
	return m.newReflectValueSingle()
}

// NewReflectValueMany creates the *[]*m.Type reflect.Value
func (m *ModelStruct) NewReflectValueMany() reflect.Value {
	return reflect.New(reflect.SliceOf(reflect.PtrTo(m.Type())))
}

func (m *ModelStruct) newReflectValueSingle() reflect.Value {
	return reflect.New(m.Type())
}

// PrimaryField returns model's primary struct field
func (m *ModelStruct) PrimaryField() *StructField {
	return m.primary
}

// PrimaryValues gets the primary values for the provided value
func (m *ModelStruct) PrimaryValues(value reflect.Value) (primaries reflect.Value, err error) {
	primaryIndex := m.primary.getFieldIndex()
	switch value.Type().Kind() {
	case reflect.Slice:
		if value.Type().Elem().Kind() != reflect.Ptr {
			err = internal.ErrUnexpectedType
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
		err = internal.ErrUnexpectedType
	}
	return
}

// RelationshipField returns the StructField for given raw field
func (m *ModelStruct) RelationshipField(field string) (*StructField, bool) {
	return StructRelField(m, field)
}

// RelationshipFields return structfields that are matched as relatinoships
func (m *ModelStruct) RelationshipFields() (rels []*StructField) {
	for _, rel := range m.relationships {
		rels = append(rels, rel)
	}
	return rels
}

// RepositoryName returns the repository name for given model
func (m *ModelStruct) RepositoryName() string {
	return m.Config().RepositoryName
}

// SchemaName returns model's schema name
func (m *ModelStruct) SchemaName() string {
	return m.schemaName
}

// SetConfig sets the config for given ModelStruct
func (m *ModelStruct) SetConfig(cfg *config.ModelConfig) error {
	log.Debugf("Setting Config: %v for model: %s", cfg, m.Collection())
	m.cfg = cfg

	if m.store == nil {
		m.store = make(map[string]interface{})
	}

	// copy the key value from the config
	for k, v := range m.cfg.Map {
		m.store[k] = v
	}

	return nil
}

// SetRepositoryName sets the repositoryName
func (m *ModelStruct) SetRepositoryName(repo string) {
	m.cfg.RepositoryName = repo
}

// SetSchemaName sets the schema name for the given model
func (m *ModelStruct) SetSchemaName(schema string) {
	m.schemaName = schema
}

// SortScopeCount returns the count of the sort fieldsb
func (m *ModelStruct) SortScopeCount() int {
	return m.sortScopeCount
}

// StoreSet sets into the store the value 'value' for given 'key'
func (m *ModelStruct) StoreSet(key string, value interface{}) {
	if m.store == nil {
		m.store = make(map[string]interface{})
	}

	log.Debugf("[STORE][%s] Set Key: %s, Value: %v", m.collectionType, key, value)
	m.store[key] = value
}

// StoreGet gets the value from the store at the key: 'key'.
func (m *ModelStruct) StoreGet(key string) (interface{}, bool) {
	if m.store == nil {
		m.store = make(map[string]interface{})
	}
	v, ok := m.store[key]
	return v, ok
}

// StoreDelete deletes the store's value at key
func (m *ModelStruct) StoreDelete(key string) {
	if m.store == nil {
		return
	}
	delete(m.store, key)
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

// Type returns model's reflect.Type
func (m ModelStruct) Type() reflect.Type {
	return m.modelType
}

// UseI18n returns the bool if the model struct uses i18n.
func (m ModelStruct) UseI18n() bool {
	return m.language != nil
}

/**

PRIVATE METHODS

*/

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

func (m *ModelStruct) initComputeSortedFields() {
	for _, sField := range m.fields {
		if sField != nil && sField.canBeSorted() {
			m.sortScopeCount++
		}
	}
}

func (m *ModelStruct) initComputeThisIncludedCount() {
	m.StoreSet(thisIncludedCountKey, len(m.relationships))
}

func initComputeNestedIncludedCount(m *ModelStruct, level, maxNestedRelLevel int) int {
	var nestedCount int
	if level != 0 {
		thisIncludedCount, _ := m.StoreGet(thisIncludedCountKey)
		nestedCount += thisIncludedCount.(int)
	}

	for _, relationship := range m.relationships {
		if level < maxNestedRelLevel {
			nestedCount += initComputeNestedIncludedCount(relationship.relationship.mStruct, level+1, maxNestedRelLevel)
		}
	}

	return nestedCount
}

// computeNestedIncludedCount computes the included count for given limit
func (m *ModelStruct) computeNestedIncludedCount(limit int) {
	nestedIncludeCount := initComputeNestedIncludedCount(m, 0, limit)

	m.StoreSet(nestedIncludedCountKey, nestedIncludeCount)
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

var ctr = &counter{}

type counter struct {
	nextID int
	lock   sync.Mutex
}

func (c *counter) next() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.nextID++
	return c.nextID
}
