package mapping

import (
	"net/url"
	"reflect"
	"time"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/annotation"
	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/namer"
)

// ModelStruct is the structure definition for the imported models.
// It contains all the collection name, fields, config, store and a model type.
type ModelStruct struct {
	// modelType contain a reflect.Type information about given model
	modelType reflect.Type
	// collectionType is neuron 'type' for given model
	collectionType string
	// Primary is a neuron primary field
	primary *StructField
	// language is a field that contains the language information
	language *StructField
	// Attributes contain attribute fields
	attributes map[string]*StructField
	// Relationships contain neuron relationship type fields.
	relationships map[string]*StructField
	// fields is a container of all public fields in the given model.
	fields []*StructField
	// field that are ready for translations
	i18n []*StructField
	// ForeignKeys is a container for the foreign keys for the relationships
	foreignKeys map[string]*StructField
	// filterKeys is a container for the filter keys
	filterKeys map[string]*StructField
	// sortScopeCount is the number of sortable fields in the model
	sortScopeCount int
	isJoin         bool
	cfg            *config.ModelConfig
	store          map[interface{}]interface{}
}

// newModelStruct creates new model struct for given type.
func newModelStruct(tp reflect.Type, collection string) *ModelStruct {
	m := &ModelStruct{modelType: tp, collectionType: collection}
	m.attributes = make(map[string]*StructField)
	m.relationships = make(map[string]*StructField)
	m.foreignKeys = make(map[string]*StructField)
	m.filterKeys = make(map[string]*StructField)
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
	if !ok {
		for _, attrField := range m.attributes {
			if field == attrField.Name() {
				return attrField, true
			}
		}
		return nil, false
	}
	return s, ok
}

// CheckField checks if the field exists within given modelstruct.
func (m *ModelStruct) CheckField(field string) (sField *StructField, err errors.DetailedError) {
	// TODO: Change the name of the function.
	return m.checkField(field)
}

// Collection returns model's collection.
func (m *ModelStruct) Collection() string {
	return m.collectionType
}

// Config gets the model's defined confgi.ModelConfig.
func (m *ModelStruct) Config() *config.ModelConfig {
	return m.cfg
}

// FieldByName returns field for provided name.
// It matches both name and neuronName.
func (m *ModelStruct) FieldByName(name string) (*StructField, bool) {
	for _, field := range m.fields {
		if field.neuronName == name || field.Name() == name {
			return field, true
		}
	}
	for _, fk := range m.filterKeys {
		if fk.neuronName == name || fk.Name() == name {
			return fk, true
		}
	}
	return nil, false
}

// ForeignKey checks and returns model's foreign key field.
// The 'fk' foreign key field name may be a Neuron name or Golang StructField name.
func (m *ModelStruct) ForeignKey(fk string) (*StructField, bool) {
	s, ok := m.foreignKeys[fk]
	if !ok {
		// If no NeuronName provided, check if this isn't the struct field name
		for _, field := range m.foreignKeys {
			if field.Name() == fk {
				return field, true
			}
		}
		return nil, false
	}
	return s, true
}

// ForeignKeys return ForeignKey Structfields for the given model.
func (m *ModelStruct) ForeignKeys() (foreigns []*StructField) {
	for _, f := range m.foreignKeys {
		foreigns = append(foreigns, f)
	}
	return foreigns
}

// Fields gets the model's fields.
func (m *ModelStruct) Fields() []*StructField {
	return m.fields
}

// FieldCount gets the field number for given model.
func (m *ModelStruct) FieldCount() int {
	return len(m.attributes) + len(m.relationships)
}

// FilterKey return model's fitler key.
func (m *ModelStruct) FilterKey(fk string) (*StructField, bool) {
	s, ok := m.filterKeys[fk]
	if !ok {
		for _, field := range m.filterKeys {
			if field.Name() == fk {
				return field, true
			}
		}
		return nil, false
	}
	return s, ok
}

// HasForeignRelationships defines if the model has any foreign relationships (not a BelongsTo relationship).
func (m *ModelStruct) HasForeignRelationships() bool {
	_, ok := m.store[hasForeignRelationships]
	return ok
}

// IsAfterLister defines if the model implements query.AfterLister interface.
func (m *ModelStruct) IsAfterLister() bool {
	_, ok := m.StoreGet(afterListerKey)
	return ok
}

// IsBeforeLister defines if the model implements query.BeforeLister interface.
func (m *ModelStruct) IsBeforeLister() bool {
	_, ok := m.StoreGet(beforeListerKey)
	return ok
}

// IsJoin defines if the model is a join table for the Many2Many relationship.
func (m *ModelStruct) IsJoin() bool {
	return m.isJoin
}

// LanguageField returns model's language field.
func (m *ModelStruct) LanguageField() *StructField {
	return m.language
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
func (m *ModelStruct) NamerFunc() namer.Namer {
	v, _ := m.StoreGet(namerFuncKey)
	return v.(namer.Namer)
}

// Primary returns model's primary field StructField.
func (m *ModelStruct) Primary() *StructField {
	return m.primary
}

// RelationField gets the relationship field for the provided string
// The 'rel' relationship field name may be a Neuron or Golang StructField name.
// If the relationship field doesn't exists returns nil and false
func (m *ModelStruct) RelationField(field string) (*StructField, bool) {
	return m.relationshipField(field)
}

// RelationFields gets all model's relationship fields.
func (m *ModelStruct) RelationFields() (relations []*StructField) {
	for _, rel := range m.relationships {
		relations = append(relations, rel)
	}
	return relations
}

// SortScopeCount returns the count of the sort fields.
func (m *ModelStruct) SortScopeCount() int {
	return m.sortScopeCount
}

// StoreDelete deletes the store's value at 'key'.
func (m *ModelStruct) StoreDelete(key interface{}) {
	if m.store == nil {
		return
	}
	log.Debug2f("[STORE][%s] Delete Key: '%s'")
	delete(m.store, key)
}

// StoreGet gets the value from the store at the key: 'key'.
func (m *ModelStruct) StoreGet(key interface{}) (interface{}, bool) {
	if m.store == nil {
		m.store = make(map[interface{}]interface{})
	}
	v, ok := m.store[key]
	log.Debug3f("[STORE][%s] Get Key: '%v' - Value: '%v', ok: %v", m.collectionType, key, v, ok)
	return v, ok
}

// StoreSet sets into the store the value 'value' for given 'key'.
func (m *ModelStruct) StoreSet(key interface{}, value interface{}) {
	if m.store == nil {
		m.store = make(map[interface{}]interface{})
	}

	log.Debug2f("[STORE][%s] Set Key: %s, Value: %v", m.collectionType, key, value)
	m.store[key] = value
}

// String implements fmt.Stringer interface.
func (m *ModelStruct) String() string {
	return m.collectionType
}

// StructFields return all the StructFields used in the ModelStruct
func (m *ModelStruct) StructFields() (fields []*StructField) {
	// add primary
	fields = append(fields, m.primary)

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
	fields = append(fields, m.i18n...)

	if m.language != nil {
		// add language field
		fields = append(fields, m.language)
	}

	// add filterKey fields
	for _, f := range m.filterKeys {
		fields = append(fields, f)
	}
	return fields
}

// Type returns model's reflect.Type.
func (m *ModelStruct) Type() reflect.Type {
	return m.modelType
}

// UseI18n returns the bool if the model struct uses i18n.
func (m ModelStruct) UseI18n() bool {
	return m.language != nil
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

func (m *ModelStruct) increaseAssignedFields() {
	var assignedFields int
	v, ok := m.store[assignedFieldsKey]
	if ok {
		assignedFields = v.(int)
	}
	assignedFields++
	m.store[assignedFieldsKey] = assignedFields
}

func (m *ModelStruct) checkField(field string) (*StructField, errors.DetailedError) {
	var (
		hasAttribute, hasRelationship bool
		sField                        *StructField
	)

	sField, hasAttribute = m.attributes[field]
	if hasAttribute {
		return sField, nil
	}

	sField, hasRelationship = m.relationships[field]
	if !hasRelationship {
		err := errors.NewDetf(class.ModelFieldNotFound, "field: '%s' not found", field)
		err.SetDetailsf("Collection: '%v', does not have field: '%v'.", m.collectionType, field)
		return nil, err
	}

	return sField, nil
}

// computeNestedIncludedCount computes the included count for given limit
func (m *ModelStruct) computeNestedIncludedCount(limit int) {
	nestedIncludeCount := m.initComputeNestedIncludedCount(0, limit)

	m.StoreSet(nestedIncludedCountKey, nestedIncludeCount)
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

func (m *ModelStruct) mapFields(modelType reflect.Type, modelValue reflect.Value, index []int) (err error) {
	for i := 0; i < modelType.NumField(); i++ {
		var fieldIndex []int

		// check if field is embedded
		tField := modelType.Field(i)

		if index == nil {
			fieldIndex = []int{i}
		} else {
			fieldIndex = append(index, i)
		}

		if tField.Anonymous {
			// the field is embedded struct or ptr to struct
			nestedModelType := tField.Type
			var nestedModelValue reflect.Value

			if nestedModelType.Kind() == reflect.Ptr {
				nestedModelType = nestedModelType.Elem()
			}

			nestedModelValue = reflect.New(nestedModelType).Elem()

			if err := m.mapFields(nestedModelType, nestedModelValue, fieldIndex); err != nil {
				log.Debugf("Mapping embedded field: %s failed: %v", tField.Name, err)
				return err
			}
			continue
		}

		// don't use private fields
		if !modelValue.Field(i).CanSet() {
			log.Debugf("Field not settable: %s", modelType.Field(i).Name)
			continue
		}

		tag, hasTag := tField.Tag.Lookup(annotation.Neuron)
		if tag == "-" {
			continue
		}

		var tagValues url.Values
		structField := newStructField(tField, m)
		structField.fieldIndex = make([]int, len(fieldIndex))
		tagValues = structField.TagValues(tag)

		copy(structField.fieldIndex, fieldIndex)
		log.Debug2f("[%s] - Field: %s with tags: %s ", m.Type().Name(), tField.Name, tagValues)

		m.increaseAssignedFields()

		// Check if field contains the name
		var neuronName string
		name := tagValues.Get(annotation.Name)
		if name != "" {
			neuronName = name
		} else {
			neuronName = m.NamerFunc()(tField.Name)
		}
		structField.neuronName = neuronName

		if !hasTag {
			m.addUntaggedField(structField)
			continue
		}

		// Set field type
		values := tagValues[annotation.FieldType]
		if len(values) == 0 {
			// return errors.NewDetf(class.ModelFieldTag, "StructField.annotation.FieldType struct field tag cannot be empty. Model: %s, field: %s", modelType.Name(), tField.Name)
			m.addUntaggedField(structField)
			continue
		}

		// Set field type
		value := values[0]
		switch value {
		case annotation.Primary, annotation.ID,
			annotation.PrimaryFull, annotation.PrimaryFullS,
			annotation.PrimaryShort:
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
		case annotation.ForeignKey, annotation.ForeignKeyFull,
			annotation.ForeignKeyFullS, annotation.ForeignKeyShort:
			if err = m.setForeignKeyField(structField); err != nil {
				return err
			}
		case annotation.FilterKey:
			structField.fieldKind = KindFilterKey
			_, ok := m.FilterKey(structField.NeuronName())
			if ok {
				return errors.NewDetf(class.ModelFieldName, "duplicated filter key name: '%s' for model: '%v'", structField.NeuronName(), m.Type().Name())
			}

			m.filterKeys[structField.neuronName] = structField
		default:
			return errors.NewDetf(class.ModelFieldTag, "unknown field type: %s. Model: %s, field: %s", value, m.Type().Name(), tField.Name)
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
	structField.fieldKind = KindAttribute
	// check if no duplicates
	_, ok := m.attributes[structField.neuronName]
	if ok {
		return errors.NewDetf(class.ModelFieldName, "duplicated neuron attribute name: '%s' for model: '%v'.",
			structField.neuronName, m.modelType.Name())
	}

	m.fields = append(m.fields, structField)

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
				// disallow nested map, arrs, maps in ptr type slices
				return errors.NewDetf(class.ModelFieldType, "structField: '%s' nested type is invalid. The model doesn't allow one of slices to ptr of slices or map", structField.Name())
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
			return errors.NewDetf(class.ModelFieldType, "map can't be a base of the slice. Field: '%s'", structField.Name())
		case reflect.Array, reflect.Slice:
			// cannot use slice of ptr slices
			return errors.NewDetf(class.ModelFieldType, "ptr slice can't be the base of the Slice field. Field: '%s'", structField.Name())
		default:
		}
	default:
		if structField.IsPtr() {
			structField.setFlag(fBasePtr)
		}
	}
	m.attributes[structField.neuronName] = structField
	return nil
}

// SetConfig sets the config for given ModelStruct
func (m *ModelStruct) setConfig(cfg *config.ModelConfig) error {
	log.Debugf("Setting Config for model: %s", m.Collection())
	m.cfg = cfg

	if m.store == nil {
		m.store = make(map[interface{}]interface{})
	}

	// copy the key value from the config
	for k, v := range m.cfg.Store {
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
			sField.setFlags(cfg.Flags...)

			// check if the field is a relationship
			if !sField.isRelationship() {
				break
			}

			rel := sField.relationship
			if rel == nil {
				rel = &Relationship{}
				sField.relationship = rel
			}

			// set on error
			if cfg.Strategy == nil {
				continue
			}

			var err errors.DetailedError
			// on create
			onCreate := cfg.Strategy.OnCreate
			if onCreate.OnError != "" {
				if err = rel.onCreate.parseOnError(onCreate.OnError); err != nil {
					return err
				}
			}

			if onCreate.QueryOrder != 0 {
				rel.onCreate.QueryOrder = onCreate.QueryOrder
			}

			// on create
			onPatch := cfg.Strategy.OnPatch
			if onPatch.OnError != "" {
				if err = rel.onPatch.parseOnError(onPatch.OnError); err != nil {
					return err
				}
			}

			if onPatch.QueryOrder != 0 {
				rel.onPatch.QueryOrder = onPatch.QueryOrder
			}

			// on create
			onDelete := cfg.Strategy.OnDelete
			if onDelete.OnError != "" {
				if err = rel.onDelete.parseOnError(onDelete.OnError); err != nil {
					return err
				}
			}

			if onDelete.QueryOrder != 0 {
				rel.onDelete.QueryOrder = onDelete.QueryOrder
			}
			if onDelete.RelationChange != "" {
				err = rel.onDelete.parseOnChange(onDelete.RelationChange)
				if err != nil {
					return err
				}
			}
			break
		}
	}
	return nil
}

func (m *ModelStruct) setForeignKeyField(structField *StructField) error {
	structField.fieldKind = KindForeignKey

	// Check if already exists
	_, ok := m.ForeignKey(structField.NeuronName())
	if ok {
		return errors.NewDetf(class.ModelFieldName, "duplicated foreign key name: '%s' for model: '%v'", structField.NeuronName(), m.Type().Name())
	}

	m.fields = append(m.fields, structField)
	m.foreignKeys[structField.neuronName] = structField

	return nil
}

func (m *ModelStruct) setLanguage(f *StructField) {
	f.fieldFlags = f.fieldFlags | fLanguage
	m.language = f
}

func (m *ModelStruct) setPrimaryField(structField *StructField) error {
	structField.fieldKind = KindPrimary
	if m.primary != nil {
		return errors.NewDetf(class.ModelFieldName, "primary field is already defined for the model: '%s'", m.Type().Name())
	}
	m.primary = structField
	m.fields = append(m.fields, structField)
	return nil
}

func (m *ModelStruct) setRelationshipField(structField *StructField) error {
	// add relationField to fields
	m.fields = append(m.fields, structField)

	// set related type
	err := structField.fieldSetRelatedType()
	if err != nil {
		return err
	}

	// check duplicates
	_, ok := m.relationships[structField.neuronName]
	if ok {
		return errors.NewDetf(class.ModelFieldName, "duplicated neuron relationship field name: '%s' for model: '%v'", structField.neuronName, m.Type().Name())
	}

	// set relationship field
	m.relationships[structField.neuronName] = structField

	return nil
}

func (m *ModelStruct) relationshipField(field string) (*StructField, bool) {
	f, ok := m.relationships[field]
	if !ok {
		for _, f = range m.relationships {
			if f.Name() == field {
				return f, true
			}
		}
		return nil, false
	}
	return f, true
}

func (m *ModelStruct) untaggedFields() []*StructField {
	v, ok := m.store[untaggedFieldKey]
	if !ok {
		return nil
	}
	return v.([]*StructField)
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

var (
	untaggedFieldKey        = untaggedFieldsStore{}
	assignedFieldsKey       = assignedFieldsStore{}
	namerFuncKey            = namerStore{}
	beforeListerKey         = beforeListerStore{}
	afterListerKey          = afterListerStore{}
	hasForeignRelationships = hasForeignRelationshipsStore{}
	thisIncludedCountKey    = thisIncludedCountKeyStore{}
	nestedIncludedCountKey  = nestedIncludedCountKeyStore{}
)

type untaggedFieldsStore struct{}
type assignedFieldsStore struct{}
type namerStore struct{}
type beforeListerStore struct{}
type afterListerStore struct{}
type hasForeignRelationshipsStore struct{}
type thisIncludedCountKeyStore struct{}
type nestedIncludedCountKeyStore struct{}
