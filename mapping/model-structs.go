package mapping

import (
	"reflect"
	"strings"
	"time"

	"github.com/neuronlabs/neuron/annotation"
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
)

// ModelStruct is the structure definition for the imported models.
// It contains all the collection name, fields, config, store and a model type.
type ModelStruct struct {
	// modelType contain a reflect.Type information about given model
	modelType reflect.Type
	// collection is the name for the models collection.
	collection string
	// structFields is a container of all fields (even fields not included in the
	// neuron model) in the given model.
	structFields []*StructField
	// fields is a container for all attributes, foreign keys
	fields []*StructField

	// Primary is a neuron primary field
	primary *StructField
	// Attributes contain attribute fields
	attributes map[string]*StructField
	// ForeignKeys is a container for the foreign keys for the relationships
	foreignKeys []*StructField
	// Relationships contain neuron relationship type fields.
	relationships []*StructField
	// private fields
	privateFields []*StructField

	// sortScopeCount is the number of sortable fields in the model
	sortScopeCount int

	// flags
	hasForeignRelationships bool
	isJoin                  bool

	createdAt *StructField
	updatedAt *StructField
	deletedAt *StructField

	namerFunc Namer

	config           *config.ModelConfig
	store            map[interface{}]interface{}
	structFieldCount int
}

// newModelStruct creates new model struct for given type.
func newModelStruct(tp reflect.Type, collection string) *ModelStruct {
	m := &ModelStruct{modelType: tp, collection: collection}
	m.attributes = make(map[string]*StructField)
	m.store = map[interface{}]interface{}{}
	return m
}

// AllowClientID checks if the model allows client settable primary key values.
func (m *ModelStruct) AllowClientID() bool {
	return m.primary.allowClientID()
}

// Attribute returns the attribute for the provided ModelStruct.
// If the attribute doesn't exists returns nil field and false.
func (m *ModelStruct) Attribute(field string) (*StructField, bool) {
	s, ok := m.attributes[field]
	return s, ok
}

// Attributes returns all field's with kind - 'KindAttribute' for given model.
func (m *ModelStruct) Attributes() (attributes []*StructField) {
	if len(m.attributes) == 0 {
		return []*StructField{}
	}
	for _, field := range m.fields {
		if field.kind == KindAttribute {
			attributes = append(attributes, field)
		}
	}
	return attributes
}

// Collection returns model's collection.
func (m *ModelStruct) Collection() string {
	return m.collection
}

// Config gets the model's defined configuration.
func (m *ModelStruct) Config() *config.ModelConfig {
	return m.config
}

// CreatedAt gets the 'CreatedAt' field for the model struct.
func (m *ModelStruct) CreatedAt() (*StructField, bool) {
	return m.createdAt, m.createdAt != nil
}

// DeletedAt gets the 'DeletedAt' field for the model struct.
func (m *ModelStruct) DeletedAt() (*StructField, bool) {
	return m.deletedAt, m.deletedAt != nil
}

// FieldByName returns structField by it's 'name'. It matches both reflect.StructField.Name and NeuronName.
func (m *ModelStruct) FieldByName(name string) (*StructField, bool) {
	for _, field := range m.fields {
		if field.neuronName == name || field.Name() == name {
			return field, true
		}
	}
	return nil, false
}

// Fields gets the model's primary, attribute and foreign key fields.
func (m *ModelStruct) Fields() (fields []*StructField) {
	return m.fields
}

// ForeignKey checks and returns model's foreign key field.
// The 'fk' foreign key field name may be a Neuron name or Golang StructField name.
func (m *ModelStruct) ForeignKey(fieldName string) (foreignKey *StructField, ok bool) {
	for _, foreignKey = range m.foreignKeys {
		if foreignKey.neuronName == fieldName || foreignKey.Name() == fieldName {
			return foreignKey, true
		}
	}
	return nil, false
}

// ForeignKeys return ForeignKey struct fields for the given model.
func (m *ModelStruct) ForeignKeys() []*StructField {
	return m.foreignKeys
}

// HasForeignRelationships defines if the model has any foreign relationships (not a BelongsTo relationship).
func (m *ModelStruct) HasForeignRelationships() bool {
	return m.hasForeignRelationships
}

// IsJoin defines if the model is a join table for the Many2Many relationship.
func (m *ModelStruct) IsJoin() bool {
	return m.isJoin
}

// MaxIncludedCount gets the maximum included field number for given model.
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

// MaxIncludedDepth gets the maximum included field depth for the queries.
func (m *ModelStruct) MaxIncludedDepth() int {
	ic, ok := m.StoreGet(nestedIncludedCountKey)
	if !ok {
		return 0
	}
	return ic.(int)
}

// NamerFunc returns namer func for the given Model.
func (m *ModelStruct) NamerFunc() Namer {
	return m.namerFunc
}

// Primary returns model's primary field StructField.
func (m *ModelStruct) Primary() *StructField {
	return m.primary
}

// PrivateFields gets model's private struct fields.
func (m *ModelStruct) PrivateFields() []*StructField {
	return m.privateFields
}

// RelationByIndex gets the relation by provided 'index'.
func (m *ModelStruct) RelationByIndex(index int) (*StructField, error) {
	if index > len(m.structFields)-1 {
		return nil, errors.Newf(ClassInvalidRelationIndex, "index out of range: '%d'", index)
	}
	structField := m.structFields[index]
	switch structField.kind {
	case KindRelationshipSingle, KindRelationshipMultiple:
		return structField, nil
	default:
		return nil, errors.Newf(ClassInvalidRelationIndex, "field with index: '%d' is not a relationship", index)
	}
}

// RelationByName gets the relationship field for the provided string
// The 'rel' relationship field name may be a Neuron or Golang StructField name.
// If the relationship field doesn't exists returns nil and false
func (m *ModelStruct) RelationByName(field string) (*StructField, bool) {
	return m.relationshipField(field)
}

// RelationFields gets all model's relationship fields.
func (m *ModelStruct) RelationFields() (relations []*StructField) {
	return m.relationships
}

// SortScopeCount returns the count of the sort fields.
func (m *ModelStruct) SortScopeCount() int {
	return m.sortScopeCount
}

// StoreDelete deletes the store's value at 'key'.
func (m *ModelStruct) StoreDelete(key interface{}) {
	log.Debug3f("[STORE][%s] Delete Key: '%s'")
	delete(m.store, key)
}

// StoreGet gets the value from the store at the key: 'key'.
func (m *ModelStruct) StoreGet(key interface{}) (interface{}, bool) {
	v, ok := m.store[key]
	log.Debug3f("[STORE][%s] Get Key: '%v' - Models: '%v', ok: %v", m.collection, key, v, ok)
	return v, ok
}

// StoreSet sets into the store the value 'value' for given 'key'.
func (m *ModelStruct) StoreSet(key interface{}, value interface{}) {
	log.Debug3f("[STORE][%s] AddModel Key: %s, Models: %v", m.collection, key, value)
	m.store[key] = value
}

// String implements fmt.Stringer interface.
func (m *ModelStruct) String() string {
	return m.collection
}

// StructFields return all the StructFields used in the ModelStruct
func (m *ModelStruct) StructFields() (fields []*StructField) {
	return m.structFields
}

// StructFieldCount returns the number of struct fields.
func (m *ModelStruct) StructFieldCount() int {
	return m.structFieldCount
}

// Type returns model's reflect.Type.
func (m *ModelStruct) Type() reflect.Type {
	return m.modelType
}

// UpdatedAt gets the 'UpdatedAt' field for the model struct.
func (m *ModelStruct) UpdatedAt() (*StructField, bool) {
	return m.updatedAt, m.updatedAt != nil
}

/**

PRIVATE METHODS

*/

func (m *ModelStruct) addUntaggedField(field *StructField) {
	untaggedFields := m.untaggedFields()
	untaggedFields = append(untaggedFields, field)
	m.store[untaggedFieldKey] = untaggedFields
}

func (m *ModelStruct) assignedFields() int {
	var assignedFields int
	v, ok := m.store[assignedFieldsKey]
	if ok {
		assignedFields = v.(int)
	}
	return assignedFields
}

func (m *ModelStruct) clearInitializeStoreKeys() {
	// clear untagged fields
	delete(m.store, untaggedFieldKey)
	delete(m.store, assignedFieldsKey)
	log.Debug3f("Cleared model: '%s' internal store key", m.Collection())
}

// computeNestedIncludedCount computes the included count for given limit
func (m *ModelStruct) computeNestedIncludedCount(limit int) {
	nestedIncludeCount := m.initComputeNestedIncludedCount(0, limit)

	m.StoreSet(nestedIncludedCountKey, nestedIncludeCount)
}

func (m *ModelStruct) findTimeRelatedFields() error {
	namerFunc := m.NamerFunc()
	var err error
	_, ok := m.CreatedAt()
	if !ok {
		// try to find attribute with default created at name.
		defaultCreatedAt := namerFunc("CreatedAt")
		if createdAtField, ok := m.Attribute(defaultCreatedAt); ok {
			if err = m.setTimeRelatedField(createdAtField, fCreatedAt); err != nil {
				return err
			}
		}
	}

	_, ok = m.DeletedAt()
	if !ok {
		// try to find attribute with default created at name.
		defaultDeletedAt := namerFunc("DeletedAt")
		if deletedAtField, ok := m.Attribute(defaultDeletedAt); ok {
			if err = m.setTimeRelatedField(deletedAtField, fDeletedAt); err != nil {
				return err
			}
		}
	}

	_, ok = m.UpdatedAt()
	if !ok {
		// try to find attribute with default created at name.
		defaultUpdatedAt := namerFunc("UpdatedAt")
		if updatedAtField, ok := m.Attribute(defaultUpdatedAt); ok {
			if err = m.setTimeRelatedField(updatedAtField, fUpdatedAt); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *ModelStruct) findUntypedInvalidAttribute(fieldName string) (*StructField, bool) {
	for _, attr := range m.attributes {
		tv := attr.TagValues(annotation.Neuron)
		_, ok := tv[annotation.FieldType]
		if ok {
			continue
		}
		// If the field has no field type check if it's value is not a foreign key
		if attr.neuronName != fieldName {
			continue
		}

		// The field is not an attribute but a foreign key. Remove it from the attribute and add to foreign keys.
		delete(m.attributes, fieldName)
		delete(m.attributes, attr.fieldName())

		m.foreignKeys = append(m.foreignKeys, attr)

		attr.kind = KindForeignKey
		return attr, true
	}
	return nil, false
}

func (m *ModelStruct) increaseAssignedFields() {
	var assignedFields int
	v, ok := m.store[assignedFieldsKey]
	if ok {
		assignedFields = v.(int)
	}
	assignedFields++
	log.Debug3f("Model: '%s' assignedFields: %d", m.Collection(), assignedFields)
	m.store[assignedFieldsKey] = assignedFields
}

func (m *ModelStruct) initComputeNestedIncludedCount(level, maxNestedRelLevel int) int {
	var nestedCount int
	if level != 0 {
		thisIncludedCount, _ := m.StoreGet(thisIncludedCountKey)
		nestedCount += thisIncludedCount.(int)
	}

	for _, relationship := range m.relationships {
		if level < maxNestedRelLevel {
			nestedCount += relationship.relationship.mStruct.initComputeNestedIncludedCount(level+1, maxNestedRelLevel)
		}
	}
	return nestedCount
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

func (m *ModelStruct) initCheckFieldTypes() (err error) {
	for _, field := range m.structFields {
		if field.kind == KindUnknown {
			continue
		}
		if err = field.initCheckFieldType(); err != nil {
			return err
		}
	}
	return nil
}

func (m *ModelStruct) mapFields(modelType reflect.Type, modelValue reflect.Value, index []int) (err error) {
	for i := 0; i < modelType.NumField(); i++ {
		fieldIndex := index

		tField := modelType.Field(i)
		fieldIndex = append(fieldIndex, i)

		// Check if field is embedded.
		if tField.Anonymous {
			return errors.Newf(ClassModelDefinition, "unsupported embedded field: '%s' in model: '%s'", tField.Name, modelType.String())
		}
		structField := newStructField(tField, m)
		m.structFields = append(m.structFields, structField)
		structField.Index = make([]int, len(fieldIndex))
		copy(structField.Index, fieldIndex)

		// Check if field is private.
		if !modelValue.Field(i).CanSet() {
			log.Debugf("Field: %s added to private fields.", modelType.Field(i).Name)
			m.privateFields = append(m.privateFields, structField)
			continue
		}

		// Get the field tag and check if it is not marked to omit - '-'.
		tag, hasTag := tField.Tag.Lookup(annotation.Neuron)
		if strings.TrimSpace(tag) == "-" {
			log.Debugf("Field: '%s' has '-' tag - adding to private fields.")
			m.privateFields = append(m.privateFields, structField)
			continue
		}
		tagValues := structField.TagValues(tag)
		log.Debug2f("[%s] - Field: %s with tags: %s ", m.Type().Name(), tField.Name, tagValues)
		m.increaseAssignedFields()

		// Check if the field had it's neuron name defined.
		neuronName := tagValues.Get(annotation.Name)
		if neuronName == "" {
			neuronName = m.namerFunc(tField.Name)
		}
		structField.neuronName = neuronName

		if !hasTag {
			m.addUntaggedField(structField)
			continue
		}

		values := tagValues[annotation.FieldType]
		switch len(values) {
		case 0:
			m.addUntaggedField(structField)
			continue
		case 1:
		default:
			return errors.NewDetf(ClassModelDefinition, "model's: '%s' field: '%s' type tag contains more than one value", m.Collection(), neuronName)
		}

		// AddModel field type
		value := values[0]
		switch value {
		case annotation.Primary, annotation.ID, annotation.PrimaryFull, annotation.PrimaryFullS, annotation.PrimaryShort:
			err = m.setPrimaryField(structField)
			if err != nil {
				return err
			}
		case annotation.Relation, annotation.RelationFull:
			err = m.setRelationshipField(structField)
			if err != nil {
				return err
			}
		case annotation.Attribute, annotation.AttributeFull:
			err = m.setAttribute(structField)
			if err != nil {
				return err
			}
		case annotation.ForeignKey, annotation.ForeignKeyFull, annotation.ForeignKeyFullS, annotation.ForeignKeyShort:
			if err = m.setForeignKeyField(structField); err != nil {
				return err
			}
		default:
			return errors.NewDetf(ClassModelDefinition, "unknown field type: %s. Model: %s, field: %s", value, m.Type().Name(), tField.Name)
		}

		if err = structField.setTagValues(); err != nil {
			return err
		}
	}
	return nil
}

func (m *ModelStruct) newReflectValueSingle() reflect.Value {
	return reflect.New(m.Type())
}

func (m *ModelStruct) setAttribute(structField *StructField) error {
	structField.kind = KindAttribute
	// check if no duplicates
	_, ok := m.attributes[structField.neuronName]
	if ok {
		return errors.NewDetf(ClassModelDefinition, "duplicated neuron attribute name: '%s' for model: '%v'.",
			structField.neuronName, m.modelType.Name())
	}

	t := structField.ReflectField().Type
	if t.Kind() == reflect.Ptr {
		structField.setFlag(fPtr)
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			structField.setFlag(fTime)
		} else {
			// this case it must be a nested struct field
			structField.setFlag(fNestedStruct)

			nStruct, err := getNestedStruct(t, structField, m.NamerFunc())
			if err != nil {
				log.Debugf("structField: %v getNestedStruct failed. %v", structField.fieldName(), err)
				return err
			}
			structField.nested = nStruct
		}

		if structField.IsPtr() {
			structField.setFlag(fBasePtr)
		}
	case reflect.Map:
		structField.setFlag(fMap)
		mapElem := t.Elem()

		// isPtr is a bool that defines if the given slice Elem is a pointer type
		// flags are not set now due to the slice sliceElem possibilities.
		var isPtr bool

		// map type must have a key of type string and value of basic type
		if mapElem.Kind() == reflect.Ptr {
			isPtr = true
			mapElem = mapElem.Elem()
		}

		// Check the map 'value' kind
		switch mapElem.Kind() {
		// struct may be time or nested struct
		case reflect.Struct:
			// check if it is time
			if mapElem == reflect.TypeOf(time.Time{}) {
				structField.setFlag(fTime)
				// otherwise it must be a nested struct
			} else {
				structField.setFlag(fNestedStruct)

				nStruct, err := getNestedStruct(mapElem, structField, m.NamerFunc())
				if err != nil {
					log.Debugf("structField: %v Map field getNestedStruct failed. %v", structField.fieldName(), err)
					return err
				}
				structField.nested = nStruct
			}
			// if the value is pointer add the base flag
			if isPtr {
				structField.setFlag(fBasePtr)
			}

			// map 'value' may be a slice or array
		case reflect.Slice, reflect.Array:
			mapElem = mapElem.Elem()
			for mapElem.Kind() == reflect.Slice || mapElem.Kind() == reflect.Array {
				mapElem = mapElem.Elem()
			}

			if mapElem.Kind() == reflect.Ptr {
				structField.setFlag(fBasePtr)
				mapElem = mapElem.Elem()
			}

			switch mapElem.Kind() {
			case reflect.Struct:
				// check if it is time
				if mapElem == reflect.TypeOf(time.Time{}) {
					structField.setFlag(fTime)
					// otherwise it must be a nested struct
				} else {
					structField.setFlag(fNestedStruct)
					nStruct, err := getNestedStruct(mapElem, structField, m.NamerFunc())
					if err != nil {
						log.Debugf("structField: %v Map field getNestedStruct failed. %v", structField.fieldName(), err)
						return err
					}
					structField.nested = nStruct
				}
			case reflect.Slice, reflect.Array, reflect.Map:
				// disallow nested map, arrays, maps in ptr type slices
				return errors.NewDetf(ClassModelDefinition, "structField: '%s' nested type is invalid. The model doesn't allow one of slices to ptr of slices or map", structField.Name())
			default:
			}
		default:
			if isPtr {
				structField.setFlag(fBasePtr)
			}
		}
	case reflect.Slice, reflect.Array:
		if t.Kind() == reflect.Slice {
			structField.setFlag(fSlice)
		} else {
			structField.setFlag(fArray)
		}

		// dereference the slice
		// check the field base type
		sliceElem := t
		for sliceElem.Kind() == reflect.Slice || sliceElem.Kind() == reflect.Array {
			sliceElem = sliceElem.Elem()
		}

		if sliceElem.Kind() == reflect.Ptr {
			// add maybe slice Field Ptr
			structField.setFlag(fBasePtr)
			sliceElem = sliceElem.Elem()
		}

		switch sliceElem.Kind() {
		case reflect.Struct:
			// check if time
			if sliceElem == reflect.TypeOf(time.Time{}) {
				structField.setFlag(fTime)
			} else {
				// this should be the nested struct
				structField.setFlag(fNestedStruct)
				nStruct, err := getNestedStruct(sliceElem, structField, m.NamerFunc())
				if err != nil {
					log.Debugf("structField: %v getNestedStruct failed. %v", structField.Name(), err)
					return err
				}
				structField.nested = nStruct
			}
		case reflect.Map:
			// map should not be allow as slice nested field
			return errors.NewDetf(ClassModelDefinition, "map can't be a base of the slice. Field: '%s'", structField.Name())
		case reflect.Array, reflect.Slice:
			// cannot use slice of ptr slices
			return errors.NewDetf(ClassModelDefinition, "ptr slice can't be the base of the Slice field. Field: '%s'", structField.Name())
		default:
		}
	default:
		if structField.IsPtr() {
			structField.setFlag(fBasePtr)
		}
	}
	m.attributes[structField.neuronName] = structField
	m.attributes[structField.fieldName()] = structField
	m.fields = append(m.fields, structField)
	return nil
}

// SetConfig sets the config for given ModelStruct
func (m *ModelStruct) setConfig(cfg *config.ModelConfig) error {
	log.Debugf("Setting Config for model: %s", m.Collection())
	m.config = cfg

	if m.store == nil {
		m.store = make(map[interface{}]interface{})
	}

	// copy the key value from the config
	for k, v := range m.config.Store {
		m.store[k] = v
	}
	return nil
}

func (m *ModelStruct) setFieldsConfigs() error {
	structFields := m.StructFields()
	for field, cfg := range m.Config().Fields {
		if cfg == nil {
			continue
		}

		for _, sField := range structFields {
			if field != sField.NeuronName() && field != sField.Name() {
				continue
			}
			if err := sField.setFlags(cfg.Flags...); err != nil {
				return err
			}

			// check if the field is a relationship
			if !sField.isRelationship() {
				break
			}

			rel := sField.relationship
			if rel == nil {
				rel = &Relationship{}
				sField.relationship = rel
			}
			break
		}
	}
	return nil
}

func (m *ModelStruct) setForeignKeyField(structField *StructField) error {
	structField.kind = KindForeignKey

	// Check if given name already exists.
	for _, foreignKey := range m.foreignKeys {
		if foreignKey == structField {
			return errors.NewDetf(ClassModelDefinition, "duplicated foreign key name: '%s' for model: '%v'", structField.NeuronName(), m.Type().Name())
		}
	}
	m.foreignKeys = append(m.foreignKeys, structField)
	m.fields = append(m.fields, structField)
	return nil
}

func (m *ModelStruct) setPrimaryField(structField *StructField) error {
	structField.kind = KindPrimary
	if m.primary != nil {
		return errors.NewDetf(ClassModelDefinition, "primary field is already defined for the model: '%s'", m.Type().Name())
	}
	m.primary = structField
	m.fields = append(m.fields, structField)
	return nil
}

func (m *ModelStruct) setRelationshipField(structField *StructField) error {
	// set related type
	if err := structField.fieldSetRelatedType(); err != nil {
		return err
	}

	for _, relation := range m.relationships {
		if relation == structField {
			return errors.NewDetf(ClassModelDefinition, "duplicated neuron relationship field name: '%s' for model: '%v'", structField.neuronName, m.Type().Name())
		}
	}
	m.relationships = append(m.relationships, structField)
	return nil
}

func (m *ModelStruct) setTimeRelatedField(field *StructField, flag fieldFlag) error {
	switch flag {
	case fCreatedAt:
		if !(field.isTime() || field.isPtrTime()) {
			return errors.NewDetf(ClassModelDefinition, "created at field: '%s' is not a time.Time field", field.Name())
		}
		if field.fieldFlags.containsFlag(flag) {
			return errors.NewDetf(ClassModelDefinition, "duplicated created at field for model: '%s'", m.Type().Name())
		}
		m.createdAt = field
	case fUpdatedAt:
		if !(field.isTime() || field.isPtrTime()) {
			return errors.NewDetf(ClassModelDefinition, "updated at field: '%s' is not a time.Time field", field.Name())
		}
		if field.fieldFlags.containsFlag(flag) {
			return errors.NewDetf(ClassModelDefinition, "duplicated updated at field for model: '%s'", m.Type().Name())
		}
		m.updatedAt = field
	case fDeletedAt:
		if !field.isPtrTime() {
			return errors.NewDetf(ClassModelDefinition, "deleted at field: '%s' is not a pointer to time.Time field", field.Name())
		}
		if field.fieldFlags.containsFlag(flag) {
			return errors.NewDetf(ClassModelDefinition, "duplicated deleted at field for model: '%s'", m.Type().Name())
		}
		m.deletedAt = field
	default:
		return errors.NewDetf(ClassInternal, "invalid related field flag: '%s'", field)
	}
	field.setFlag(flag)
	return nil
}

func (m *ModelStruct) relationshipField(fieldName string) (relation *StructField, ok bool) {
	for _, relation = range m.relationships {
		if relation.neuronName == fieldName || relation.Name() == fieldName {
			return relation, true
		}
	}
	return nil, false
}

func (m *ModelStruct) untaggedFields() []*StructField {
	v, ok := m.store[untaggedFieldKey]
	if !ok {
		return nil
	}
	return v.([]*StructField)
}

var (
	untaggedFieldKey       = untaggedFieldsStore{}
	assignedFieldsKey      = assignedFieldsStore{}
	thisIncludedCountKey   = thisIncludedCountKeyStore{}
	nestedIncludedCountKey = nestedIncludedCountKeyStore{}
)

type untaggedFieldsStore struct{}
type assignedFieldsStore struct{}

type thisIncludedCountKeyStore struct{}
type nestedIncludedCountKeyStore struct{}
