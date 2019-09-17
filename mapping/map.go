package mapping

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/neuronlabs/inflection"

	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/neuron-core/annotation"
	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/namer"
)

// ModelMap contains mapped models ( as reflect.Type ) to its ModelStruct representation.
type ModelMap struct {
	models      map[reflect.Type]*ModelStruct
	collections map[string]reflect.Type
	Configs     map[string]*config.ModelConfig

	DefaultRepository string
	NamerFunc         namer.Namer

	nestedIncludedLimit int
}

// NewModelMap creates new model map with default 'namerFunc' and a controller config 'c'.
func NewModelMap(namerFunc namer.Namer, c *config.Controller) *ModelMap {
	if c.Models == nil {
		c.Models = make(map[string]*config.ModelConfig)
	}
	return &ModelMap{
		models:              make(map[reflect.Type]*ModelStruct),
		collections:         make(map[string]reflect.Type),
		DefaultRepository:   c.DefaultRepositoryName,
		NamerFunc:           namerFunc,
		Configs:             c.Models,
		nestedIncludedLimit: c.IncludedDepthLimit,
	}
}

// Get gets the *ModelStruct for the provided 'model'.
func (m *ModelMap) Get(model reflect.Type) *ModelStruct {
	return m.models[m.getType(model)]
}

// GetByCollection gets *ModelStruct by the 'collection'.
func (m *ModelMap) GetByCollection(collection string) *ModelStruct {
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
		return nil, errors.NewDetf(class.ModelNotMapped, "model: '%s' is not mapped", t.Name())
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
	var err error
	for _, model := range models {
		// build the model's structure and set into model map.
		mStruct, err := buildModelStruct(model, m.NamerFunc)
		if err != nil {
			return err
		}

		if err = m.Set(mStruct); err != nil {
			continue
		}

		var modelConfig *config.ModelConfig

		modelConfig, ok := m.Configs[mStruct.collectionType]
		if !ok {
			modelConfig = &config.ModelConfig{}
			m.Configs[mStruct.collectionType] = modelConfig
		}
		modelConfig.Collection = mStruct.Collection()

		if err := mStruct.setConfig(modelConfig); err != nil {
			log.Errorf("Setting config for model: '%s' failed.", mStruct.Collection())
			return err
		}

		if modelConfig.RepositoryName == "" {
			// if the model implements repository Name
			repositoryNamer, ok := model.(RepositoryNamer)
			if ok {
				modelConfig.RepositoryName = repositoryNamer.RepositoryName()
			}
		}
		mStruct.namerFunc = m.NamerFunc
	}

	for _, modelStruct := range m.models {
		if err = m.setUntaggedFields(modelStruct); err != nil {
			return err
		}

		if modelStruct.assignedFields() == 0 {
			err = errors.NewDetf(class.ModelMappingNoFields, "model: '%s' have no fields defined", modelStruct.Type().Name())
			return err
		}

		if modelStruct.primary == nil {
			err = errors.NewDetf(class.ModelMappingNoFields, "model: '%s' have no primary field type defined", modelStruct.Type().Name())
			return err
		}
		if err = modelStruct.setFieldsConfigs(); err != nil {
			return err
		}
	}

	for _, model := range m.models {
		if err = m.setModelRelationships(model); err != nil {
			return err
		}

		if err = model.initCheckFieldTypes(); err != nil {
			return err
		}
		model.initComputeSortedFields()
		m.setByCollection(model)

		if err = model.findTimeRelatedFields(); err != nil {
			return err
		}
	}

	m.computeNestedIncludedCount()
	if err = m.setRelationships(); err != nil {
		return err
	}

	for _, model := range m.models {
		model.structFieldCount = len(model.StructFields())
	}
	return nil
}

// Set sets the *ModelStruct for given map.
// If the model already exists the function returns an error.
func (m *ModelMap) Set(value *ModelStruct) error {
	_, ok := m.models[value.modelType]
	if ok {
		return errors.NewDetf(class.ModelInSchemaAlreadyRegistered, "Model: %s already registered", value.Type())
	}

	_, ok = m.collections[value.collectionType]
	if ok {
		return errors.NewDetf(class.ModelInSchemaAlreadyRegistered, "Model: %s already registered", value.Type())
	}

	m.models[value.modelType] = value
	m.collections[value.collectionType] = value.Type()

	return nil
}

// setByCollection sets the model by it's collection.
func (m *ModelMap) setByCollection(ms *ModelStruct) {
	m.collections[ms.Collection()] = ms.Type()
}

// computeNestedIncludedCount computes the limits for the nested included count for each model.
func (m *ModelMap) computeNestedIncludedCount() {
	limit := m.nestedIncludedLimit

	for _, model := range m.models {
		model.initComputeThisIncludedCount()
	}

	for _, model := range m.models {
		modelLimit := limit
		if model.cfg.IncludedDepthLimit != nil {
			modelLimit = *model.cfg.IncludedDepthLimit
		}
		model.computeNestedIncludedCount(modelLimit)
	}
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

func (m *ModelMap) setModelRelationships(model *ModelStruct) error {
	for _, relField := range model.RelationFields() {
		relType := relField.getRelatedModelType()

		val := m.Get(relType)
		if val == nil {
			return errors.NewDetf(class.InternalModelRelationNotMapped, "model: %v, not precalculated but is used in relationships for: %v field in %v model", relType, relField.Name(), model.Type().Name())
		}

		relField.setRelatedModel(val)

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
				case name2:
					rel.joinModel = other
				case rel.joinModelName:
					rel.joinModel = other
				default:
					continue
				}
			}

			if rel.joinModel == nil {
				return errors.NewDetf(class.ModelRelationshipJoinModel, "Join Model not found for model: '%s' relationship: '%s'", model.Type().Name(), relField.Name())
			}

			rel.joinModel.isJoin = true
		}
	}
	return nil
}

func (m *ModelMap) setRelationships() error {
	for _, model := range m.models {
		for _, relField := range model.RelationFields() {
			// relationship gets the relationship between the fields
			relationship := relField.Relationship()

			// get structfield neuron tags
			tags := relField.TagValues(relField.ReflectField().Tag.Get(annotation.Neuron))

			// get proper foreign key field name
			fkeyFieldNames := tags[annotation.ForeignKey]
			log.Debugf("Relationship field: %s, foreign key name: %s", relField.Name(), fkeyFieldNames)
			// check field type
			var (
				foreignKey, m2mForeignKey string
				taggedForeign             bool
			)

			switch len(fkeyFieldNames) {
			case 0:
			case 1:
				foreignKey = fkeyFieldNames[0]
				taggedForeign = true
			case 2:
				foreignKey = fkeyFieldNames[0]
				if foreignKey == "_" {
					foreignKey = ""
				} else {
					taggedForeign = true
				}

				m2mForeignKey = fkeyFieldNames[1]
				if m2mForeignKey == "_" {
					m2mForeignKey = ""
				}
			default:
				log.Errorf("Too many foreign key tag values for the relationship field: '%s' in model: '%s' ", relField.Name(), model.Type().Name())
				return errors.NewDet(class.ModelRelationshipForeign, "relationship field tag 'foreign key' has too values")
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
						return errors.NewDetf(class.ModelFieldForeignKeyNotFound, "Foreign key: '%s' not found for the relationship: '%s'. Model: '%s'", foreignKey, relField.Name(), model.Type().Name())
					}
					// the primary field type of the model should match current's model type.
					if model.Primary().ReflectField().Type != fk.ReflectField().Type {
						log.Errorf("the foreign key in model: %v for the to-many relation: '%s' doesn't match the primary field type. Wanted: %v, Is: %v", modelWithFK.Type().Name(), relField.Name(), model.Type().Name(), model.primary.ReflectField().Type, fk.ReflectField().Type)
						return errors.NewDetf(class.ModelRelationshipForeign, "foreign key type doesn't match the primary field type of the root model")
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
						return errors.NewDetf(class.ModelFieldForeignKeyNotFound, "Related Model Foreign Key: '%s' not found for the relationship: '%s'. Model: '%s'", m2mForeignKeyName, relField.Name(), model.Type().Name())
					}

					// the primary field type of the model should match current's model type.
					if relationship.mStruct.primary.ReflectField().Type != m2mFK.ReflectField().Type {
						log.Debugf("the foreign key of the related model: '%v' for the many-to-many relation: '%s' doesn't match the primary field type. Wanted: %v, Is: %v", relationship.mStruct.Type().Name(), relField.Name(), model.Type().Name(), model.primary.reflectField.Type, fk.ReflectField().Type)
						return errors.NewDetf(class.ModelRelationshipForeign, "foreign key type doesn't match the primary field type of the root model")
					}

					relationship.mtmRelatedForeignKey = m2mFK
				} else {
					relationship.setKind(RelHasMany)
					if foreignKey == "" {
						// the foreign key for any of the slice relationships should be
						// model's that contains the relationship name concantated with the 'ID'.
						foreignKey = relField.Name() + "ID"
					}

					modelWithFK := relationship.mStruct
					// get the name from the NamerFunc.
					fkeyName := m.NamerFunc(foreignKey)

					// check if given FK exists in the model's definitions.
					fk, ok := modelWithFK.ForeignKey(fkeyName)
					if !ok {
						foreignKey = model.modelType.Name() + "ID"
						fk, ok = modelWithFK.ForeignKey(m.NamerFunc(foreignKey))
						if !ok {
							log.Errorf("Foreign key: '%s' not found within Model: '%s'", fkeyName, modelWithFK.Type().Name())
							return errors.NewDetf(class.ModelFieldForeignKeyNotFound, "Foreign key: '%s' not found for the relationship: '%s'. Model: '%s'", fkeyName, relField.Name(), model.Type().Name())
						}
					}
					// the primary field type of the model should match current's model type.
					if model.Primary().ReflectField().Type != fk.ReflectField().Type {
						log.Debugf("the foreign key in model: %v for the to-many relation: '%s' doesn't match the primary field type. Wanted: %v, Is: %v", modelWithFK.Type().Name(), relField.Name(), model.Type().Name(), model.Primary().ReflectField().Type, fk.ReflectField().Type)
						return errors.NewDetf(class.ModelRelationshipForeign, "foreign key type doesn't match the primary field type of the root model")
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
				if ok {
					relationship.setKind(RelBelongsTo)
					relationship.foreignKey = fk
					continue
				}

				// if the foreign key is not found it must be a has one model or an invalid field name was provided
				if !taggedForeign {
					// check if the model might have a name of belong's to
					modelsForeign := relField.relationship.mStruct.Type().Name() + "ID"
					// check if the foreign key would be the name of the related structure
					relatedTypeName := m.NamerFunc(modelsForeign)
					fk, ok = model.ForeignKey(relatedTypeName)
					if !ok {
						fk, ok = model.findUntypedInvalidAttribute(relatedTypeName)
					}
					if ok {
						relationship.kind = RelBelongsTo
						relationship.foreignKey = fk
						continue
					}
				}

				fk, ok = model.findUntypedInvalidAttribute(fkeyName)
				if ok {
					relationship.setKind(RelBelongsTo)
					relationship.foreignKey = fk
					continue
				}

				// if none of the foreign were found within the 'model', try to find it within
				// related model. It would be then a 'HasOne' relationship
				fk, ok = relationship.mStruct.ForeignKey(fkeyName)
				if ok {
					relationship.kind = RelHasOne
					// set the foreign key for the given relationship
					relationship.foreignKey = fk
					continue
				}

				fk, ok = relationship.mStruct.findUntypedInvalidAttribute(fkeyName)
				if ok {
					relationship.kind = RelHasOne
					relationship.foreignKey = fk
					continue
				}

				modelsForeign := m.NamerFunc(relField.mStruct.Type().Name() + "ID")

				fk, ok = relationship.mStruct.ForeignKey(modelsForeign)
				if ok {
					relationship.kind = RelHasOne
					// set the foreign key for the given relationship
					relationship.foreignKey = fk
					continue
				}

				fk, ok = relationship.mStruct.findUntypedInvalidAttribute(modelsForeign)
				if ok {
					relationship.kind = RelHasOne
					// set the foreign key for the given relationship
					relationship.foreignKey = fk
					continue
				}

				// provided invalid foreign field name
				return errors.NewDetf(class.ModelFieldForeignKeyNotFound, "foreign key: '%s' not found for the relationship: '%s'. Model: '%s'", fkeyName, relField.Name(), model.Type().Name())

			}
		}

		for _, relField := range model.RelationFields() {
			if relField.relationship.isMany2Many() {
				continue
			}

			// check if the foreign key match related primary field type
			var match bool
			switch relField.relationship.kind {
			case RelBelongsTo:
				if relField.relationship.mStruct.Primary().ReflectField().Type == relField.relationship.foreignKey.ReflectField().Type {
					match = true
				}
			case RelHasMany, RelHasOne:
				if relField.mStruct.Primary().ReflectField().Type == relField.relationship.foreignKey.ReflectField().Type {
					match = true
				}
			}

			if !match {
				log.Errorf("the foreign key in model: %v for the belongs-to relation: %s with model: %s is of invalid type. Wanted: %v, Is: %v", model,
					relField.relationship.foreignKey.mStruct.Type().Name(),
					relField.relationship.modelType.Name(),
					relField.relationship.Struct().Primary().ReflectField().Type,
					relField.relationship.foreignKey.ReflectField().Type)
				return errors.NewDetf(class.ModelRelationshipForeign, "foreign key: '%s' doesn't match model's: '%s' primary key type", relField.relationship.foreignKey.Name(), relField.relationship.mStruct.Type().Name())
			}
		}
	}

	for _, model := range m.models {
		for _, relField := range model.relationships {
			if relField.Relationship().kind != RelBelongsTo {
				model.hasForeignRelationships = true
				return nil
			}
		}
	}

	return nil
}

func (m *ModelMap) setUntaggedFields(model *ModelStruct) (err error) {
	untaggedFields := model.untaggedFields()
	if untaggedFields == nil {
		return nil
	}

	for _, field := range untaggedFields {
		// if there is no struct field tag and the field's ToLower name  is 'id'
		// set it as the model's primary key.
		if strings.ToLower(field.Name()) == "id" {
			err = model.setPrimaryField(field)
			if err != nil {
				return err
			}
			continue
		}

		// if strings.ToLower(field.Name())
		otherModel := m.Get(field.BaseType())
		if otherModel != nil {
			// set them as the relationships
			err = model.setRelationshipField(field)
			if err != nil {
				return err
			}
			continue
		}

		if strings.HasSuffix(field.Name(), "ID") {
			var isForeignKey bool
			for otherType := range m.models {
				if strings.HasPrefix(field.Name(), otherType.Name()) {
					isForeignKey = true
					break
				}
			}
			if isForeignKey {
				if err = model.setForeignKeyField(field); err != nil {
					return err
				}
				continue
			}
		}

		if err = model.setAttribute(field); err != nil {
			return err
		}
	}

	for _, field := range untaggedFields {
		if err = field.setTagValues(); err != nil {
			return err
		}
	}
	return nil
}

// buildModelStruct builds the model struct for the provided model with the given namer function
func buildModelStruct(model interface{}, namerFunc namer.Namer) (modelStruct *ModelStruct, err error) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		err = errors.NewDetf(class.ModelMappingInvalidType, `provided model in invalid format. The model must be of struct or ptr type, but is: %v`, modelType)
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

	modelStruct = newModelStruct(modelType, collection)
	modelStruct.namerFunc = namerFunc
	// Define the function definition

	// map fields
	if err := modelStruct.mapFields(modelType, modelValue, nil); err != nil {
		return nil, err
	}

	if ptrValue.MethodByName("BeforeList").IsValid() {
		modelStruct.isBeforeLister = true
	}
	if ptrValue.MethodByName("AfterList").IsValid() {
		modelStruct.isAfterLister = true
	}
	if ptrValue.MethodByName("BeforeCount").IsValid() {
		modelStruct.isBeforeCounter = true
	}
	if ptrValue.MethodByName("AfterCount").IsValid() {
		modelStruct.isAfterCounter = true
	}

	return modelStruct, nil
}

func getNestedStruct(t reflect.Type, sFielder StructFielder, namerFunc namer.Namer) (*NestedStruct, error) {
	nestedStruct := newNestedStruct(t, sFielder)
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

		nestedField := newNestedField(nestedStruct, sFielder, nField)

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
				case annotation.Name:
					nestedField.structField.neuronName = tValue[0]
				case annotation.FieldType:
					if tValue[0] != annotation.NestedField {
						log.Debugf("Invalid annotation.NestedField value: '%s' for field: %s", tValue[0], nestedField.structField.Name())
						return nil, errors.NewDetf(class.ModelFieldType, "provided field type: '%s' is not allowed for the nested struct field: '%s'", nestedField.structField.reflectField.Type, nestedField.structField.Name())
					}
				case annotation.Flags:
					if err := nestedField.structField.setFlags(tValue...); err != nil {
						return nil, err
					}
				}
			}
		}

		if nestedField.structField.NeuronName() == "" {
			nestedField.structField.neuronName = namerFunc(nField.Name)
		}

		switch nestedField.structField.NeuronName() {
		case "relationships", "links":
			return nil, errors.NewDetf(class.ModelFieldName, "nested field within: '%s' field in the model: '%s' has forbidden Neuron name: '%s'",
				nestedStruct.Attr().Name(),
				nestedStruct.Attr().Struct().Type().Name(),
				nestedField.structField.NeuronName(),
			)
		}

		if _, ok = NestedStructSubField(nestedStruct, nestedField.structField.NeuronName()); ok {
			return nil, errors.NewDetf(class.ModelFieldName, "nestedStruct: %v already has one nestedField: '%s'. The fields must be uniquely named", nestedStruct.Type().Name(), nestedField.structField.Name())
		}
		NestedStructSetSubfield(nestedStruct, nestedField)

		nFType := nField.Type
		if nFType.Kind() == reflect.Ptr {
			nFType = nFType.Elem()
			nestedField.structField.setFlag(fPtr)
		}

		switch nFType.Kind() {
		case reflect.Struct:
			if nFType == reflect.TypeOf(time.Time{}) {
				nestedField.structField.setFlag(fTime)
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

			if nestedField.structField.IsPtr() {
				nestedField.structField.setFlag(fBasePtr)
			}
		case reflect.Map:
			nestedField.structField.setFlag(fMap)
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
					nestedField.structField.setFlag(fTime)
					// otherwise it must be a nested struct
				} else {
					nestedField.structField.setFlag(fNestedStruct)

					nStruct, err := getNestedStruct(mapElem, nestedField, namerFunc)
					if err != nil {
						log.Debugf("nestedField: %v Map field getNestedStruct failed. %v", nestedField.structField.fieldName(), err)
						return nil, err
					}

					nestedField.structField.nested = nStruct
				}
				// if the value is pointer add the base flag
				if isPtr {
					nestedField.structField.setFlag(fBasePtr)
				}
			case reflect.Slice, reflect.Array:
				mapElem = mapElem.Elem()
				for mapElem.Kind() == reflect.Slice || mapElem.Kind() == reflect.Array {
					mapElem = mapElem.Elem()
				}

				if mapElem.Kind() == reflect.Ptr {
					nestedField.structField.setFlag(fBasePtr)
					mapElem = mapElem.Elem()
				}

				switch mapElem.Kind() {
				case reflect.Struct:
					// check if it is time
					if mapElem == reflect.TypeOf(time.Time{}) {
						nestedField.structField.setFlag(fTime)
						// otherwise it must be a nested struct
					} else {
						nestedField.structField.setFlag(fNestedStruct)

						nStruct, err := getNestedStruct(mapElem, nestedField, namerFunc)
						if err != nil {
							log.Debugf("nestedField: %v Map field getNestedStruct failed. %v", nestedField.structField.fieldName(), err)
							return nil, err
						}
						nestedField.structField.nested = nStruct
					}
				case reflect.Slice, reflect.Array, reflect.Map:
					// disallow nested map, arrs, maps in ptr type slices
					return nil, errors.NewDetf(class.ModelFieldType, "structField: '%s' nested type is invalid. The model doesn't allow one of slices to ptr of slices or map", nestedField.structField.Name())
				default:
				}
			}
		case reflect.Slice, reflect.Array:
			if nFType.Kind() == reflect.Slice {
				nestedField.structField.setFlag(fSlice)
			} else {
				nestedField.structField.setFlag(fArray)
			}
			for nFType.Kind() == reflect.Slice || nFType.Kind() == reflect.Array {
				nFType = nFType.Elem()
			}

			if nFType.Kind() == reflect.Ptr {
				nestedField.structField.setFlag(fBasePtr)
				nFType = nFType.Elem()
			}

			switch nFType.Kind() {
			case reflect.Struct:
				// check if time
				if nFType == reflect.TypeOf(time.Time{}) {
					nestedField.structField.setFlag(fTime)
				} else {
					// this should be the nested struct
					nestedField.structField.setFlag(fNestedStruct)
					nStruct, err := getNestedStruct(nFType, nestedField, namerFunc)
					if err != nil {
						log.Debugf("nestedField: %v getNestedStruct failed. %v", nestedField.structField.fieldName(), err)
						return nil, err
					}
					nestedField.structField.nested = nStruct
				}
			case reflect.Slice, reflect.Ptr, reflect.Map, reflect.Array:
				return nil, errors.NewDetf(class.ModelFieldType, "nested field can't be a slice of pointer to slices|map|arrays. NestedField: '%s' within NestedStruct:'%s'", nestedField.structField.Name(), nestedStruct.modelType.Name())
			}
		default:
			// basic type (ints, uints, string, floats)
			// do nothing

			if nestedField.structField.IsPtr() {
				nestedField.structField.setFlag(fBasePtr)
			}
		}

		var tagValue = nestedField.structField.NeuronName()
		if nestedField.structField.isOmitEmpty() {
			tagValue += ",omitempty"
		}

		marshalField.Tag = reflect.StructTag(fmt.Sprintf(`json:"%s"`, tagValue))
		marshalFields = append(marshalFields, marshalField)
	}

	NestedStructSetMarshalType(nestedStruct, reflect.StructOf(marshalFields))

	return nestedStruct, nil
}
