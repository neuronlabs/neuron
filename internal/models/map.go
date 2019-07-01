package models

import (
	"fmt"
	"net/url"
	"reflect"
	"sync"
	"time"
	"unicode"

	"github.com/jinzhu/inflection"

	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/errors/class"
	"github.com/neuronlabs/neuron/log"

	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/namer"
)

// ModelMap contains mapped models ( as reflect.Type ) to its ModelStruct representation.
// Allow concurrent safe gets and sets to the map.
type ModelMap struct {
	sync.RWMutex
	models      map[reflect.Type]*ModelStruct
	collections map[string]reflect.Type
	Configs     map[string]*config.ModelConfig

	DefaultRepository string
	NamerFunc         namer.Namer
}

// NewModelMap creates new model map
func NewModelMap(namerFunc namer.Namer, c *config.Controller) *ModelMap {
	if c.Models == nil {
		c.Models = make(map[string]*config.ModelConfig)
	}

	var modelMap = &ModelMap{
		models:            make(map[reflect.Type]*ModelStruct),
		collections:       make(map[string]reflect.Type),
		DefaultRepository: c.DefaultRepositoryName,
		NamerFunc:         namerFunc,
		Configs:           c.Models,
	}

	return modelMap
}

// ComputeNestedIncludedCount computes the limits for the nested included count for each model in each schema
func (m *ModelMap) ComputeNestedIncludedCount(limit int) {
	for _, model := range m.models {
		model.initComputeThisIncludedCount()
	}

	for _, model := range m.models {
		model.computeNestedIncludedCount(limit)
	}
}

// Get is concurrent safe getter of model structs.
func (m *ModelMap) Get(model reflect.Type) *ModelStruct {
	m.RLock()
	defer m.RUnlock()

	return m.models[m.getType(model)]
}

// GetByCollection gets model by collection name
func (m *ModelMap) GetByCollection(collection string) *ModelStruct {
	m.RLock()
	defer m.RUnlock()
	t, ok := m.collections[collection]
	if !ok || t == nil {
		return nil
	}
	return m.models[t]
}

// GetModelStruct gets the model from the model map.
func (m *ModelMap) GetModelStruct(model interface{}) (*ModelStruct, error) {
	mStruct, ok := model.(*ModelStruct)
	if ok {
		return mStruct, nil
	}
	t := reflect.TypeOf(model)
	mStruct = m.Get(t)
	if mStruct == nil {
		return nil, errors.Newf(class.ModelNotMapped, "model: '%s' is not mapped", t.Name())
	}
	return mStruct, nil
}

// Models returns all models set within given model map.
func (m *ModelMap) Models() []*ModelStruct {
	structs := []*ModelStruct{}

	for _, model := range m.models {
		structs = append(structs, model)
	}
	return structs
}

// ModelByName gets the model by it's struct name.
func (m *ModelMap) ModelByName(name string) *ModelStruct {
	for _, model := range m.models {
		if model.Type().Name() == name {
			return model
		}
	}
	return nil
}

// RegisterModels registers the model within the model map container.
func (m *ModelMap) RegisterModels(models ...interface{}) error {
	// iterate over models and register one by one
	for _, model := range models {
		// build the model's structure and set into schema's model map
		mStruct, err := buildModelStruct(model, m.NamerFunc)
		if err != nil {
			return err
		}

		if err := m.Set(mStruct); err != nil {
			continue
		}

		var modelConfig *config.ModelConfig

		modelConfig, ok := m.Configs[mStruct.collectionType]
		if !ok {
			modelConfig = &config.ModelConfig{}
			m.Configs[mStruct.collectionType] = modelConfig
		}
		modelConfig.Collection = mStruct.Collection()

		if err := mStruct.SetConfig(modelConfig); err != nil {
			log.Errorf("Setting config for model: '%s' failed.", mStruct.Collection())
			return err
		}

		if modelConfig.RepositoryName == "" {
			// if the model implements repository Name
			repositoryNamer, ok := model.(namer.RepositoryNamer)
			if ok {
				modelConfig.RepositoryName = repositoryNamer.RepositoryName()
			}
		}

		mStruct.StoreSet(namerFuncKey, m.NamerFunc)
	}

	for _, model := range m.models {
		if err := m.setModelRelationships(model); err != nil {
			return err
		}

		if err := model.initCheckFieldTypes(); err != nil {
			return err
		}
		model.initComputeSortedFields()

		m.SetByCollection(model)
	}

	err := m.setRelationships()
	if err != nil {
		return err
	}

	return nil
}

// Set sets the modelstruct for given map
// If the model already exists the function returns an error
func (m *ModelMap) Set(value *ModelStruct) error {
	m.Lock()
	defer m.Unlock()

	_, ok := m.models[value.modelType]
	if ok {
		return errors.Newf(class.ModelInSchemaAlreadyRegistered, "Model: %s already registered", value.Type())
	}

	_, ok = m.collections[value.collectionType]
	if ok {
		return errors.Newf(class.ModelInSchemaAlreadyRegistered, "Model: %s already registered", value.Type())
	}

	m.models[value.modelType] = value
	m.collections[value.collectionType] = value.Type()

	return nil
}

// SetByCollection sets the model by it's collection
func (m *ModelMap) SetByCollection(ms *ModelStruct) {
	m.Lock()
	defer m.Unlock()

	m.collections[ms.Collection()] = ms.Type()
}

func (m *ModelMap) setModelRelationships(model *ModelStruct) error {

	for _, relField := range model.RelationshipFields() {
		relType := FieldsRelatedModelType(relField)

		val := m.Get(relType)
		if val == nil {
			return errors.Newf(class.InternalModelRelationNotMapped, "model: %v, not precalculated but is used in relationships for: %v field in %v model", relType, relField.FieldName(), model.Type().Name())
		}

		relField.SetRelatedModel(val)

		// if the field was typed as many2many searc hfor it's join model
		if relField.relationship.Kind() == RelMany2Many {
			// look for backreference field and join model
			rel := relField.relationship

			var name1, name2 string
			if rel.joinModelName == "" {
				name1 = model.Type().Name() + inflection.Plural(rel.modelType.Name())
				name2 = rel.modelType.Name() + inflection.Plural(model.modelType.Name())
			}

			for _, other := range m.models {
				switch other.Type().Name() {
				case name1:
					rel.joinModel = other
					break
				case name2:
					rel.joinModel = other
					break
				case rel.joinModelName:
					rel.joinModel = other
					break
				default:
					continue
				}
			}

			if rel.joinModel == nil {
				return errors.Newf(class.ModelRelationshipJoinModel, "Join Model not found for model: '%s' relationship: '%s'", model.Type().Name(), relField.Name())
			}

			rel.joinModel.isJoin = true
		}
	}
	return nil
}

func (m *ModelMap) setRelationships() error {
	// iterate over all models from schema
	for _, model := range m.models {
		for _, relField := range model.RelationshipFields() {
			// relationship gets the relationship between the fields
			relationship := relField.Relationship()

			// get structfield jsonapi tags
			tags := relField.TagValues(relField.ReflectField().Tag.Get(internal.AnnotationNeuron))

			// get proper foreign key field name
			fkeyFieldNames := tags[internal.AnnotationForeignKey]
			log.Debugf("Relationship field: %s, foreign key name: %s", relField.Name(), fkeyFieldNames)
			// check field type
			var foreignKey, m2mForeignKey string
			switch len(fkeyFieldNames) {
			case 0:
			case 1:
				foreignKey = fkeyFieldNames[0]
			case 2:
				foreignKey = fkeyFieldNames[0]
				if foreignKey == "_" {
					foreignKey = ""
				}

				m2mForeignKey = fkeyFieldNames[1]
				if m2mForeignKey == "_" {
					m2mForeignKey = ""
				}
			default:
				log.Errorf("Too many foreign key tag values for the relationship field: '%s' in model: '%s' ", relField.Name(), model.Type().Name())
				return errors.New(class.ModelRelationshipForeign, "relationship field tag 'foreign key' has too values")
			}
			log.Debugf("ForeignKeys: %v", fkeyFieldNames)

			switch relField.ReflectField().Type.Kind() {
			case reflect.Slice:
				if relationship.isMany2Many() {
					// check if foreign key has it's name
					if foreignKey == "" {
						foreignKey = model.modelType.Name() + "ID"
					}

					// get the foreign keys from the join model
					modelWithFK := relationship.joinModel

					// get the name from the NamerFunc.
					fkeyName := m.NamerFunc(foreignKey)

					// check if given FK exists in the model's definitions.
					fk, ok := modelWithFK.ForeignKey(fkeyName)
					if !ok {
						log.Errorf("Foreign key: '%s' not found within Model: '%s'", foreignKey, modelWithFK.Type().Name())
						return errors.Newf(class.ModelFieldForeignKeyNotFound, "Foreign key: '%s' not found for the relationship: '%s'. Model: '%s'", foreignKey, relField.Name(), model.Type().Name())
					}
					// the primary field type of the model should match current's model type.
					if model.PrimaryField().ReflectField().Type != fk.ReflectField().Type {
						log.Errorf("the foreign key in model: %v for the to-many relation: '%s' doesn't match the primary field type. Wanted: %v, Is: %v", modelWithFK.Type().Name(), relField.Name(), model.Type().Name(), model.PrimaryField().ReflectField().Type, fk.ReflectField().Type)
						return errors.Newf(class.ModelRelationshipForeign, "foreign key type doesn't match the primary field type of the root model")
					}
					relationship.setForeignKey(fk)

					// check if m2mForeignKey is set
					if m2mForeignKey == "" {
						m2mForeignKey = relationship.modelType.Name() + "ID"
					}

					// get the name from the NamerFunc.
					m2mForeignKeyName := m.NamerFunc(m2mForeignKey)

					// check if given FK exists in the model's definitions.
					m2mFK, ok := modelWithFK.ForeignKey(m2mForeignKeyName)
					if !ok {
						log.Debugf("Foreign key: '%s' not found within Model: '%s'", fkeyName, modelWithFK.Type().Name())
						return errors.Newf(class.ModelFieldForeignKeyNotFound, "Related Model Foreign Key: '%s' not found for the relationship: '%s'. Model: '%s'", m2mForeignKeyName, relField.Name(), model.Type().Name())
					}

					// the primary field type of the model should match current's model type.
					if relationship.mStruct.PrimaryField().ReflectField().Type != m2mFK.ReflectField().Type {
						log.Debugf("the foreign key of the related model: '%v' for the many-to-many relation: '%s' doesn't match the primary field type. Wanted: %v, Is: %v", relationship.mStruct.Type().Name(), relField.Name(), model.Type().Name(), model.PrimaryField().ReflectField().Type, fk.ReflectField().Type)
						return errors.Newf(class.ModelRelationshipForeign, "foreign key type doesn't match the primary field type of the root model")
					}

					relationship.mtmRelatedForeignKey = m2mFK
				} else {
					relationship.setKind(RelHasMany)
					if foreignKey == "" {
						// the foreign key for any of the slice relationships should be
						// model's that contains the relationship name concantated with the 'ID'.
						foreignKey = model.modelType.Name() + "ID"
					}
					modelWithFK := relationship.mStruct
					// get the name from the NamerFunc.
					fkeyName := m.NamerFunc(foreignKey)

					// check if given FK exists in the model's definitions.
					fk, ok := modelWithFK.ForeignKey(fkeyName)
					if !ok {
						log.Errorf("Foreign key: '%s' not found within Model: '%s'", fkeyName, modelWithFK.Type().Name())
						return errors.Newf(class.ModelFieldForeignKeyNotFound, "Foreign key: '%s' not found for the relationship: '%s'. Model: '%s'", fkeyName, relField.Name(), model.Type().Name())
					}
					// the primary field type of the model should match current's model type.
					if model.PrimaryField().ReflectField().Type != fk.ReflectField().Type {
						log.Debugf("the foreign key in model: %v for the to-many relation: '%s' doesn't match the primary field type. Wanted: %v, Is: %v", modelWithFK.Type().Name(), relField.Name(), model.Type().Name(), model.PrimaryField().ReflectField().Type, fk.ReflectField().Type)
						return errors.Newf(class.ModelRelationshipForeign, "foreign key type doesn't match the primary field type of the root model")
					}
					relationship.setForeignKey(fk)
				}
			case reflect.Ptr, reflect.Struct:
				// check if it is belongs_to or has_one relationship
				if foreignKey == "" {
					// if foreign key has no given name the default value
					// is the relationship field name concantated with 'ID'
					foreignKey = relField.ReflectField().Name + "ID"
				}

				// use the NamerFunc to get the field's name
				fkeyName := m.NamerFunc(foreignKey)

				// search for the foreign key within the given model
				fk, ok := model.ForeignKey(fkeyName)
				if !ok {
					// if the foreign key is not found it must be a has one model or an invalid field name was provided
					relationship.setKind(RelHasOne)

					fk, ok = relationship.mStruct.ForeignKey(fkeyName)
					if !ok {
						// provided invalid foreign field name
						return errors.Newf(class.ModelFieldForeignKeyNotFound, "foreign key: '%s' not found for the relationship: '%s'. Model: '%s'", fkeyName, relField.Name(), model.Type().Name())
					}

					// then the check if the current model's primary field is of the same type as the foreign key
					if model.PrimaryField().ReflectField().Type != fk.ReflectField().Type {
						log.Errorf("foreign key in model: %v for the has-one relation: %s within model: %s is of invalid type. Wanted: %v, Is: %v",
							fk.Struct().Type().Name(),
							relField.Name(),
							model.Type().Name(),
							model.PrimaryField().ReflectField().Type,
							fk.ReflectField().Type)
						return errors.New(class.ModelRelationshipForeign, "foreign key type doesn't match model with has one relationship primary key type")
					}
				} else {
					// relationship is of BelongsTo kind
					// check if the related model struct primary field is of the same type as the given foreing key
					if relationship.mStruct.PrimaryField().ReflectField().Type != fk.ReflectField().Type {
						log.Errorf("the foreign key in model: %v for the belongs-to relation: %s with model: %s is of invalid type. Wanted: %v, Is: %v", model,
							relField.Name(),
							FieldsRelatedModelType(relField).Name(),
							FieldsRelatedModelStruct(relField).PrimaryField().ReflectField().Type,
							fk.ReflectField().Type)
						return errors.New(class.ModelRelationshipForeign, "foreign key type doesn't match model's with belongs to relationship primary key type")
					}
					relationship.kind = RelBelongsTo
				}
				// set the foreign key for the given relationship
				relationship.foreignKey = fk
			}
		}
	}

modelLoop:
	for _, model := range m.models {
		for _, relField := range model.relationships {
			if relField.Relationship().kind != RelBelongsTo {
				model.StoreSet(hasForeignRelationships, struct{}{})
				break modelLoop
			}
		}
	}

	return nil
}

func (m *ModelMap) getType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}
	return t
}

func (m *ModelMap) getSimilarCollections(collection string) (simillar []string) {
	/**

	TO IMPLEMENT:

	find closest match collection

	*/
	return []string{}
}

// buildModelStruct builds the model struct for the provided model with the given namer function
func buildModelStruct(
	model interface{},
	namerFunc namer.Namer,
) (*ModelStruct, error) {
	var err error

	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		err = errors.Newf(class.ModelMappingInvalidType, `provided model in invalid format. The model must be of struct or ptr type, but is: %v`, modelType)
		return nil, err
	}

	// check and set the interfaces
	ptrValue := reflect.New(modelType)
	modelValue := reflect.New(modelType).Elem()

	var collection string

	collectioner, ok := model.(Collectioner)
	if ok {
		collection = collectioner.CollectionName()
	} else {
		collection = namerFunc(inflection.Plural(modelType.Name()))
	}

	modelStruct := newModelStruct(modelType, collection)

	// Define the function definition
	var (
		mapFields      func(modelType reflect.Type, modelValue reflect.Value, index []int) error
		assignedFields int
	)

	init := true

	// assign the function to it
	mapFields = func(modelType reflect.Type, modelValue reflect.Value, index []int) error {

		for i := 0; i < modelType.NumField(); i++ {
			var fieldIndex []int

			// check if field is embedded
			tField := modelType.Field(i)

			if init {
				fieldIndex = []int{i}
				init = false
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

				if err := mapFields(nestedModelType, nestedModelValue, fieldIndex); err != nil {
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

			tag, ok := tField.Tag.Lookup(internal.AnnotationNeuron)
			if !ok {
				continue
			}

			if tag == "-" {
				continue
			}

			var tagValues url.Values

			structField := NewStructField(tField, modelStruct)
			tagValues = structField.TagValues(tag)

			structField.fieldIndex = make([]int, len(fieldIndex))
			copy(structField.fieldIndex, fieldIndex)

			log.Debugf("[%s] - Field: %s with tags: %s ", modelStruct.Type().Name(), tField.Name, tagValues)

			assignedFields++

			// Check if field contains the name
			var neuronName string
			name := tagValues.Get(internal.AnnotationName)
			if name != "" {
				neuronName = name
			} else {
				neuronName = namerFunc(tField.Name)
			}

			FieldsSetNeuronName(structField, neuronName)

			// Set field type
			values := tagValues[internal.AnnotationFieldType]
			if len(values) == 0 {
				return errors.Newf(class.ModelFieldTag, "StructField.annotationFieldType struct field tag cannot be empty. Model: %s, field: %s", modelType.Name(), tField.Name)
			}

			// Set field type
			value := values[0]
			switch value {
			case internal.AnnotationPrimary, internal.AnnotationID,
				internal.AnnotationPrimaryFull, internal.AnnotationPrimaryFullS:
				structField.fieldKind = KindPrimary
				modelStruct.primary = structField
				modelStruct.fields = append(modelStruct.fields, structField)
			case internal.AnnotationRelation, internal.AnnotationRelationFull:
				// add relationField to fields
				modelStruct.fields = append(modelStruct.fields, structField)

				// set related type
				err = structField.fieldSetRelatedType()
				if err != nil {
					return err
				}

				// check duplicates
				_, ok := modelStruct.RelationshipField(neuronName)
				if ok {
					return errors.Newf(class.ModelFieldName, "duplicated jsonapi relationship field name: '%s' for model: '%v'", neuronName, modelType.Name())
				}

				// set relationship field
				modelStruct.relationships[structField.neuronName] = structField
			case internal.AnnotationAttribute, internal.AnnotationAttributeFull:
				structField.fieldKind = KindAttribute
				// check if no duplicates
				_, ok := modelStruct.attributes[neuronName]
				if ok {
					return errors.Newf(class.ModelFieldName, "duplicated jsonapi attribute name: '%s' for model: '%v'.",
						structField.neuronName, modelStruct.modelType.Name())
				}

				modelStruct.fields = append(modelStruct.fields, structField)

				t := structField.ReflectField().Type
				if t.Kind() == reflect.Ptr {
					FieldSetFlag(structField, FPtr)
					t = t.Elem()
				}

				switch t.Kind() {
				case reflect.Struct:
					if t == reflect.TypeOf(time.Time{}) {
						FieldSetFlag(structField, FTime)
					} else {
						// this case it must be a nested struct field
						FieldSetFlag(structField, FNestedStruct)

						nStruct, err := getNestedStruct(t, structField, namerFunc)
						if err != nil {
							log.Debugf("structField: %v getNestedStruct failed. %v", structField.fieldName(), err)
							return err
						}
						structField.nested = nStruct
					}

					if FieldIsPtr(structField) {
						FieldSetFlag(structField, FBasePtr)
					}
				case reflect.Map:
					FieldSetFlag(structField, FMap)

					mapElem := t.Elem()

					// isPtr is a bool that defines if the given slice Elem is a pointer type
					// flags are not set now due to the slice sliceElem possiblities
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
							FieldSetFlag(structField, FTime)
							// otherwise it must be a nested struct
						} else {
							FieldSetFlag(structField, FNestedStruct)

							nStruct, err := getNestedStruct(mapElem, structField, namerFunc)
							if err != nil {
								log.Debugf("structField: %v Map field getNestedStruct failed. %v", structField.fieldName(), err)
								return err
							}
							structField.nested = nStruct
						}
						// if the value is pointer add the base flag
						if isPtr {
							FieldSetFlag(structField, FBasePtr)
						}

						// map 'value' may be a slice or array
					case reflect.Slice, reflect.Array:
						mapElem = mapElem.Elem()
						for mapElem.Kind() == reflect.Slice || mapElem.Kind() == reflect.Array {
							mapElem = mapElem.Elem()
						}

						if mapElem.Kind() == reflect.Ptr {
							FieldSetFlag(structField, FBasePtr)
							mapElem = mapElem.Elem()
						}

						switch mapElem.Kind() {
						case reflect.Struct:
							// check if it is time
							if mapElem == reflect.TypeOf(time.Time{}) {
								FieldSetFlag(structField, FTime)
								// otherwise it must be a nested struct
							} else {
								FieldSetFlag(structField, FNestedStruct)
								nStruct, err := getNestedStruct(mapElem, structField, namerFunc)
								if err != nil {
									log.Debugf("structField: %v Map field getNestedStruct failed. %v", structField.fieldName(), err)
									return err
								}
								structField.nested = nStruct
							}
						case reflect.Slice, reflect.Array, reflect.Map:
							// disallow nested map, arrs, maps in ptr type slices
							return errors.Newf(class.ModelFieldType, "structField: '%s' nested type is invalid. The model doesn't allow one of slices to ptr of slices or map", structField.Name())
						default:
						}
					default:
						if isPtr {
							FieldSetFlag(structField, FBasePtr)
						}

					}
				case reflect.Slice, reflect.Array:
					if t.Kind() == reflect.Slice {
						FieldSetFlag(structField, FSlice)
					} else {
						FieldSetFlag(structField, FArray)
					}

					// dereference the slice
					// check the field base type
					sliceElem := t
					for sliceElem.Kind() == reflect.Slice || sliceElem.Kind() == reflect.Array {
						sliceElem = sliceElem.Elem()
					}

					if sliceElem.Kind() == reflect.Ptr {
						// add maybe slice Field Ptr
						FieldSetFlag(structField, FBasePtr)
						sliceElem = sliceElem.Elem()
					}

					switch sliceElem.Kind() {
					case reflect.Struct:
						// check if time
						if sliceElem == reflect.TypeOf(time.Time{}) {
							FieldSetFlag(structField, FTime)
						} else {
							// this should be the nested struct
							FieldSetFlag(structField, FNestedStruct)
							nStruct, err := getNestedStruct(sliceElem, structField, namerFunc)
							if err != nil {
								log.Debugf("structField: %v getNestedStruct failed. %v", structField.Name(), err)
								return err
							}
							structField.nested = nStruct
						}
					case reflect.Map:
						// map should not be allow as slice nested field
						return errors.Newf(class.ModelFieldType, "map can't be a base of the slice. Field: '%s'", structField.Name())
					case reflect.Array, reflect.Slice:
						// cannot use slice of ptr slices
						return errors.Newf(class.ModelFieldType, "ptr slice can't be the base of the Slice field. Field: '%s'", structField.Name())
					default:
					}
				default:
					if FieldIsPtr(structField) {
						FieldSetFlag(structField, FBasePtr)
					}

				}
				modelStruct.attributes[structField.neuronName] = structField
			case internal.AnnotationForeignKey, internal.AnnotationForeignKeyFull,
				internal.AnnotationForeignKeyFullS:
				structField.fieldKind = KindForeignKey

				// Check if already exists
				_, ok := modelStruct.ForeignKey(structField.NeuronName())
				if ok {
					return errors.Newf(class.ModelFieldName, "duplicated foreign key name: '%s' for model: '%v'", structField.NeuronName(), modelStruct.Type().Name())
				}

				modelStruct.fields = append(modelStruct.fields, structField)
				modelStruct.foreignKeys[structField.neuronName] = structField
			case internal.AnnotationFilterKey:
				structField.fieldKind = KindFilterKey
				_, ok := modelStruct.FilterKey(structField.NeuronName())
				if ok {
					return errors.Newf(class.ModelFieldName, "duplicated filter key name: '%s' for model: '%v'", structField.NeuronName(), modelStruct.Type().Name())
				}

				modelStruct.filterKeys[structField.neuronName] = structField
			default:
				return errors.Newf(class.ModelFieldTag, "unknown field type: %s. Model: %s, field: %s", value, modelStruct.Type().Name(), tField.Name)
			}

			// iterate over structfield tags
			for key, values := range tagValues {
				switch key {
				case internal.AnnotationFieldType, internal.AnnotationName:
					continue
				case internal.AnnotationFlags:
					for _, value := range values {

						switch value {
						case internal.AnnotationClientID:
							FieldSetFlag(structField, FClientID)
						case internal.AnnotationNoFilter:
							FieldSetFlag(structField, FNoFilter)
						case internal.AnnotationHidden:
							FieldSetFlag(structField, FHidden)
						case internal.AnnotationNotSortable:
							FieldSetFlag(structField, FSortable)
						case internal.AnnotationISO8601:
							FieldSetFlag(structField, FIso8601)
						case internal.AnnotationOmitEmpty:
							FieldSetFlag(structField, FOmitempty)
						case internal.AnnotationI18n:
							FieldSetFlag(structField, FI18n)
							modelStruct.i18n = append(modelStruct.i18n, structField)
						case internal.AnnotationLanguage:
							modelStruct.setLanguage(structField)

						}
					}
				case internal.AnnotationManyToMany:
					if !structField.isRelationship() {
						log.Debugf("Field: %s tagged with: %s is not a relationship.", structField.ReflectField().Name, internal.AnnotationManyToMany)
						return errors.New(class.ModelFieldTag, "many2many tag on non relationship field")
					}
					r := structField.relationship
					if r == nil {
						r = &Relationship{}
						structField.relationship = r
					}

					r.kind = RelMany2Many

					// first value is join model
					// the second is the backreference field
					switch len(values) {
					case 1:
						if values[0] != "_" {
							r.joinModelName = values[0]
						}
					case 0:
					default:
						err = errors.New(class.ModelFieldTag, "relationship many2many tag has too many values")
						return err
					}
				}
			}
		}
		return nil
	}

	// map fields
	if err := mapFields(modelType, modelValue, nil); err != nil {
		return nil, err
	}

	if assignedFields == 0 {
		err = errors.Newf(class.ModelMappingNoFields, "model: '%s' have no fields defined", modelType.Name())
		return nil, err
	}

	if modelStruct.primary == nil {
		err = errors.Newf(class.ModelMappingNoFields, "model: '%s' have no primary field type defined", modelType.Name())
		return nil, err
	}

	if ptrValue.MethodByName("BeforeList").IsValid() {
		modelStruct.StoreSet(beforeListerKey, struct{}{})
	}

	if ptrValue.MethodByName("AfterList").IsValid() {
		modelStruct.StoreSet(afterListerKey, struct{}{})
	}

	return modelStruct, nil
}

func getNestedStruct(
	t reflect.Type, sFielder StructFielder, namerFunc namer.Namer,
) (*NestedStruct, error) {
	nestedStruct := NewNestedStruct(t, sFielder)
	v := reflect.New(t).Elem()
	var marshalFields []reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		nField := t.Field(i)
		vField := v.Field(i)

		marshalField := reflect.StructField{
			Name: nField.Name,
			Type: nField.Type,
		}

		// should get the field's name
		if unicode.IsLower(rune(nField.Name[0])) || !vField.CanSet() {
			marshalField.Tag = `json:"-"`
			// marshalFields = append(marshalFields, marshalField)
			continue
		}

		nestedField := NewNestedField(nestedStruct, sFielder, nField)

		tag, ok := nField.Tag.Lookup("neuron")
		if ok {
			if tag == "-" {
				marshalField.Tag = reflect.StructTag(`json:"-"`)
				marshalFields = append(marshalFields, marshalField)
				continue
			}

			// get the tag values
			tagValues := nestedField.structField.TagValues(tag)
			for tKey, tValue := range tagValues {
				switch tKey {
				case internal.AnnotationName:
					FieldsSetNeuronName(nestedField.structField, tValue[0])
				case internal.AnnotationFieldType:
					if tValue[0] != internal.AnnotationNestedField {
						log.Debugf("Invalid annotationNestedField value: '%s' for field: %s", tValue[0], nestedField.structField.Name())
						return nil, errors.Newf(class.ModelFieldType, "provided field type: '%s' is not allowed for the nested struct field: '%s'", nestedField.structField.FieldType(), nestedField.structField.Name())
					}
				case internal.AnnotationFlags:
					for _, value := range tValue {
						switch value {
						case internal.AnnotationNoFilter:
							FieldSetFlag(nestedField.structField, FNoFilter)
						case internal.AnnotationHidden:
							FieldSetFlag(nestedField.structField, FHidden)
						case internal.AnnotationNotSortable:
							FieldSetFlag(nestedField.structField, FSortable)
						case internal.AnnotationISO8601:
							FieldSetFlag(nestedField.structField, FIso8601)
						case internal.AnnotationOmitEmpty:
							FieldSetFlag(nestedField.structField, FOmitempty)
						}

					}
				}
			}
		}

		if nestedField.structField.NeuronName() == "" {
			FieldsSetNeuronName(nestedField.structField, namerFunc(nField.Name))
		}

		switch nestedField.structField.NeuronName() {
		case "relationships", "links":
			return nil, errors.Newf(class.ModelFieldName, "nested field within: '%s' field in the model: '%s' has forbidden API name: '%s'",
				NestedStructAttr(nestedStruct).Name(),
				FieldsStruct(NestedStructAttr(nestedStruct)).Type().Name(),
				nestedField.structField.NeuronName(),
			)
		default:

		}

		if _, ok = NestedStructSubField(nestedStruct, nestedField.structField.NeuronName()); ok {
			return nil, errors.Newf(class.ModelFieldName, "nestedStruct: %v already has one nestedField: '%s'. The fields must be uniquely named", nestedStruct.Type().Name(), nestedField.structField.Name())
		}
		NestedStructSetSubfield(nestedStruct, nestedField)

		nFType := nField.Type
		if nFType.Kind() == reflect.Ptr {
			nFType = nFType.Elem()
			FieldSetFlag(nestedField.structField, FPtr)
		}

		switch nFType.Kind() {
		case reflect.Struct:
			if nFType == reflect.TypeOf(time.Time{}) {
				FieldSetFlag(nestedField.structField, FTime)
			} else {
				// nested nested field
				nStruct, err := getNestedStruct(nFType, nestedField, namerFunc)
				if err != nil {
					log.Debug("NestedField: %s. getNestedStruct failed. %v", nField.Name, err)
					return nil, err
				}

				nestedField.structField.nested = nStruct

				marshalField.Type = nStruct.marshalType
			}

			if FieldIsPtr(nestedField.structField) {
				FieldSetFlag(nestedField.structField, FBasePtr)
			}
		case reflect.Map:
			FieldSetFlag(nestedField.structField, FMap)
			// should it be disallowed?
			// check the inner kind
			mapElem := nFType.Elem()

			var isPtr bool
			if mapElem.Kind() == reflect.Ptr {
				isPtr = true
				mapElem = mapElem.Elem()
			}

			switch mapElem.Kind() {
			case reflect.Struct:

				// check if it is time
				if mapElem == reflect.TypeOf(time.Time{}) {
					FieldSetFlag(nestedField.structField, FTime)
					// otherwise it must be a nested struct
				} else {
					FieldSetFlag(nestedField.structField, FNestedStruct)

					nStruct, err := getNestedStruct(mapElem, nestedField, namerFunc)
					if err != nil {
						log.Debugf("nestedField: %v Map field getNestedStruct failed. %v", nestedField.structField.fieldName(), err)
						return nil, err
					}

					nestedField.structField.nested = nStruct
				}
				// if the value is pointer add the base flag
				if isPtr {
					FieldSetFlag(nestedField.structField, FBasePtr)
				}
			case reflect.Slice, reflect.Array:
				mapElem = mapElem.Elem()
				for mapElem.Kind() == reflect.Slice || mapElem.Kind() == reflect.Array {
					mapElem = mapElem.Elem()
				}

				if mapElem.Kind() == reflect.Ptr {
					FieldSetFlag(nestedField.structField, FBasePtr)
					mapElem = mapElem.Elem()
				}

				switch mapElem.Kind() {
				case reflect.Struct:
					// check if it is time
					if mapElem == reflect.TypeOf(time.Time{}) {
						FieldSetFlag(nestedField.structField, FTime)
						// otherwise it must be a nested struct
					} else {
						FieldSetFlag(nestedField.structField, FNestedStruct)

						nStruct, err := getNestedStruct(mapElem, nestedField, namerFunc)
						if err != nil {
							log.Debugf("nestedField: %v Map field getNestedStruct failed. %v", nestedField.structField.fieldName(), err)
							return nil, err
						}
						nestedField.structField.nested = nStruct
					}
				case reflect.Slice, reflect.Array, reflect.Map:
					// disallow nested map, arrs, maps in ptr type slices
					return nil, errors.Newf(class.ModelFieldType, "structField: '%s' nested type is invalid. The model doesn't allow one of slices to ptr of slices or map", nestedField.structField.Name())
				default:
				}
			}

		case reflect.Slice, reflect.Array:
			if nFType.Kind() == reflect.Slice {
				FieldSetFlag(nestedField.structField, FSlice)
			} else {
				FieldSetFlag(nestedField.structField, FArray)
			}
			for nFType.Kind() == reflect.Slice || nFType.Kind() == reflect.Array {
				nFType = nFType.Elem()
			}

			if nFType.Kind() == reflect.Ptr {
				FieldSetFlag(nestedField.structField, FBasePtr)
				nFType = nFType.Elem()
			}

			switch nFType.Kind() {
			case reflect.Struct:
				// check if time
				if nFType == reflect.TypeOf(time.Time{}) {
					FieldSetFlag(nestedField.structField, FTime)
				} else {
					// this should be the nested struct
					FieldSetFlag(nestedField.structField, FNestedStruct)
					nStruct, err := getNestedStruct(nFType, nestedField, namerFunc)
					if err != nil {
						log.Debugf("nestedField: %v getNestedStruct failed. %v", nestedField.structField.fieldName(), err)
						return nil, err
					}
					nestedField.structField.nested = nStruct
				}
			case reflect.Slice, reflect.Ptr, reflect.Map, reflect.Array:
				return nil, errors.Newf(class.ModelFieldType, "nested field can't be a slice of pointer to slices|map|arrays. NestedField: '%s' within NestedStruct:'%s'", nestedField.structField.Name(), nestedStruct.modelType.Name())
			}

		default:
			// basic type (ints, uints, string, floats)
			// do nothing

			if FieldIsPtr(nestedField.structField) {
				FieldSetFlag(nestedField.structField, FBasePtr)
			}
		}

		var tagValue = nestedField.structField.NeuronName()

		if FieldIsOmitEmpty(nestedField.structField) {
			tagValue += ",omitempty"
		}

		marshalField.Tag = reflect.StructTag(fmt.Sprintf(`json:"%s"`, tagValue))
		marshalFields = append(marshalFields, marshalField)
	}

	NestedStructSetMarshalType(nestedStruct, reflect.StructOf(marshalFields))

	return nestedStruct, nil
}
