package mapping

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/neuronlabs/inflection"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
)

// ModelMap contains mapped models ( as reflect.Type ) to its ModelStruct representation.
type ModelMap struct {
	collections map[string]*ModelStruct

	// DefaultNotNull defines if the non pointer fields are by default database not nulls defined.
	Options             *MapOptions
	DefaultRepository   string
	nestedIncludedLimit int
}

// MapOptions are the options for the model map.
type MapOptions struct {
	DefaultNotNull bool
	ModelNotNull   map[Model]struct{}
	Namer          NamingConvention
}

// NewModelMap creates new model map with default 'namerFunc' and a controller config 'c'.
func NewModelMap(o *MapOptions) *ModelMap {
	return &ModelMap{
		Options:     o,
		collections: make(map[string]*ModelStruct),
	}
}

// GetByCollection gets *ModelStruct by the 'collection'.
func (m *ModelMap) GetByCollection(collection string) (*ModelStruct, bool) {
	model, ok := m.collections[collection]
	return model, ok
}

// GetModelStruct gets the model from the model map.
func (m *ModelMap) GetModelStruct(model Model) (*ModelStruct, bool) {
	mStruct, ok := m.collections[model.NeuronCollectionName()]
	return mStruct, ok
}

// MustModelStruct gets the model from the model map.
func (m *ModelMap) MustModelStruct(model Model) *ModelStruct {
	mStruct, ok := m.collections[model.NeuronCollectionName()]
	if !ok {
		panic("model not found")
	}
	return mStruct
}

// Models returns all models set within given model map.
func (m *ModelMap) Models() []*ModelStruct {
	var models []*ModelStruct
	for _, model := range m.collections {
		models = append(models, model)
	}
	return models
}

// ModelByName gets the model by it's struct name.
func (m *ModelMap) ModelByName(name string) *ModelStruct {
	for _, model := range m.collections {
		if model.Type().Name() == name {
			return model
		}
	}
	return nil
}

// RegisterModels registers the model within the model map container.
func (m *ModelMap) RegisterModels(models ...Model) error {
	// iterate over models and register one by one
	var thisModels []*ModelStruct
	for _, model := range models {
		log.Debug3f("Registering Model: '%T'", model)
		// check if the model wasn't already registered within the ModelMap 'm'.
		if m.collections[model.NeuronCollectionName()] != nil {
			return errors.Wrapf(ErrModelContainer, "model: '%s' was already registered.", model.NeuronCollectionName())
		}
		// build the model's structure and set into model map.
		mStruct, err := buildModelStruct(model, m.Options.Namer)
		if err != nil {
			return err
		}

		if err = m.addModel(mStruct); err != nil {
			return err
		}
		thisModels = append(thisModels, mStruct)
		mStruct.namer = m.Options.Namer
	}

	var err error
	for _, modelStruct := range thisModels {
		if err = m.setUntaggedFields(modelStruct); err != nil {
			return err
		}

		if modelStruct.assignedFields() == 0 {
			log.Debug3f("Assigned fields: %v - number: %v", modelStruct.assignedFields(), modelStruct.store[assignedFieldsKey])
			err = errors.WrapDetf(ErrModelDefinition, "model: '%s' have no fields defined", modelStruct.Type().Name())
			return err
		}

		if modelStruct.primary == nil {
			err = errors.WrapDetf(ErrModelDefinition, "model: '%s' have no primary field type defined", modelStruct.Type().Name())
			return err
		}
	}

	for _, model := range m.collections {
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

	for _, model := range thisModels {
		model.structFieldCount = len(model.StructFields())
		model.clearInitializeStoreKeys()
	}

	// Extract database and codec indexes.
	for _, model := range thisModels {
		defaultNotNull := m.Options.DefaultNotNull
		for md := range m.Options.ModelNotNull {
			mStruct, ok := m.GetModelStruct(md)
			if !ok {
				continue
			}
			if mStruct == model {
				defaultNotNull = true
				break
			}
		}
		if err = model.extractDatabaseTags(defaultNotNull); err != nil {
			return err
		}
		if err = model.extractCodecTags(); err != nil {
			return err
		}
	}
	return nil
}

// addModel sets the *ModelStruct for given map.
// If the model already exists the function returns an error.
func (m *ModelMap) addModel(mStruct *ModelStruct) error {
	// Check if collection with given name already exists.
	_, ok := m.collections[mStruct.collection]
	if ok {
		return errors.WrapDetf(ErrModelContainer, "Model: %s already registered", mStruct.Type())
	}
	m.collections[mStruct.collection] = mStruct
	return nil
}

// setByCollection sets the model by it's collection.
func (m *ModelMap) setByCollection(ms *ModelStruct) {
	m.collections[ms.Collection()] = ms
}

// computeNestedIncludedCount computes the limits for the nested included count for each model.
func (m *ModelMap) computeNestedIncludedCount() {
	limit := m.nestedIncludedLimit

	for _, model := range m.collections {
		model.initComputeThisIncludedCount()
	}

	for _, model := range m.collections {
		modelLimit := limit
		model.computeNestedIncludedCount(modelLimit)
	}
}

func (m *ModelMap) setModelRelationships(model *ModelStruct) error {
	for _, relField := range model.RelationFields() {
		relType := relField.getRelatedModelType()
		relationModel, ok := reflect.New(relType).Interface().(Model)
		if !ok {
			return errors.Wrapf(ErrModelDefinition, "relation field: '%s' for model: '%s' type doesn't implement Model interface", relField.Name(), model.Type().Name())
		}
		relatedModelStruct, ok := m.GetModelStruct(relationModel)
		if !ok {
			fmt.Printf("RelationModel: %T\n", relationModel)
			return errors.WrapDetf(ErrInternal, "model: %v, not defined but is used in relationships for: %v field in %v model", relType, relField.Name(), model.Type().Name())
		}

		relField.setRelatedModel(relatedModelStruct)

		// if the field was typed as many2many search for it's join model
		if relField.relationship.Kind() == RelMany2Many {
			// look for backreference field and join model
			rel := relField.relationship

			var name1, name2 string
			if rel.joinModelName == "" {
				name1 = model.Type().Name() + inflection.Plural(rel.modelType.Name())
				name2 = rel.modelType.Name() + inflection.Plural(model.modelType.Name())
			}

			for _, other := range m.collections {
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
				return errors.WrapDetf(ErrModelDefinition, "Join Model not found for model: '%s' relationship: '%s'", model.Type().Name(), relField.Name())
			}

			rel.joinModel.isJoin = true
		}
	}
	return nil
}

func (m *ModelMap) setRelationships() error {
	for _, model := range m.collections {
		for _, relField := range model.RelationFields() {
			// relationship gets the relationship between the fields
			relationship := relField.Relationship()

			// get struct field neuron tags
			tags := relField.TagValues(relField.ReflectField().Tag.Get(AnnotationNeuron))

			// get proper foreign key field name
			foreignKeyFieldNames := tags[AnnotationForeignKey]
			log.Debugf("Relationship field: %s, foreign key name: %s", relField.Name(), foreignKeyFieldNames)
			// check field type
			var (
				foreignKey, m2mForeignKey string
				taggedForeign             bool
			)

			switch len(foreignKeyFieldNames) {
			case 0:
			case 1:
				foreignKey = foreignKeyFieldNames[0]
				taggedForeign = true
			case 2:
				foreignKey = foreignKeyFieldNames[0]
				if foreignKey == "_" {
					foreignKey = ""
				} else {
					taggedForeign = true
				}

				m2mForeignKey = foreignKeyFieldNames[1]
				if m2mForeignKey == "_" {
					m2mForeignKey = ""
				}
			default:
				log.Errorf("Too many foreign key tag values for the relationship field: '%s' in model: '%s' ", relField.Name(), model.Type().Name())
				return errors.WrapDet(ErrModelDefinition, "relationship field tag 'foreign key' has too values")
			}
			log.Debugf("ForeignKeys: %v", foreignKeyFieldNames)

			switch relField.ReflectField().Type.Kind() {
			case reflect.Slice:
				if relationship.isMany2Many() {
					// check if foreign key has it's name
					if foreignKey == "" {
						foreignKey = model.modelType.Name() + "ID"
					}

					// get the foreign keys from the join model
					modelWithFK := relationship.joinModel

					// get the name from the NamingConvention.
					foreignKeyName := m.Options.Namer.Namer(foreignKey)

					// check if given FK exists in the model's definitions.
					fk, ok := modelWithFK.ForeignKey(foreignKeyName)
					if !ok {
						log.Errorf("Foreign key: '%s' not found within Model: '%s'", foreignKey, modelWithFK.Type().Name())
						return errors.WrapDetf(ErrModelDefinition, "Foreign key: '%s' not found for the relationship: '%s'. Model: '%s'", foreignKey, relField.Name(), model.Type().Name())
					}
					// the primary field type of the model should match current's model type.
					if model.Primary().ReflectField().Type != fk.ReflectField().Type {
						log.Errorf("the foreign key in model: %v for the to-many relation: '%s' doesn't match the primary field type. Wanted: %v, Is: %v", modelWithFK.Type().Name(), relField.Name(), model.Type().Name(), model.primary.ReflectField().Type, fk.ReflectField().Type)
						return errors.WrapDetf(ErrModelDefinition, "foreign key type doesn't match the primary field type of the root model")
					}
					relationship.setForeignKey(fk)

					// check if m2mForeignKey is set
					if m2mForeignKey == "" {
						m2mForeignKey = relationship.modelType.Name() + "ID"
					}

					// get the name from the NamingConvention.
					m2mForeignKeyName := m.Options.Namer.Namer(m2mForeignKey)

					// check if given FK exists in the model's definitions.
					m2mFK, ok := modelWithFK.ForeignKey(m2mForeignKeyName)
					if !ok {
						log.Debugf("Foreign key: '%s' not found within Model: '%s'", foreignKeyName, modelWithFK.Type().Name())
						return errors.WrapDetf(ErrModelDefinition, "Related Model Foreign Key: '%s' not found for the relationship: '%s'. Model: '%s'", m2mForeignKeyName, relField.Name(), model.Type().Name())
					}

					// the primary field type of the model should match current's model type.
					if relationship.mStruct.primary.ReflectField().Type != m2mFK.ReflectField().Type {
						log.Debugf("the foreign key of the related model: '%v' for the many-to-many relation: '%s' doesn't match the primary field type. Wanted: %v, Is: %v", relationship.mStruct.Type().Name(), relField.Name(), model.Type().Name(), model.primary.reflectField.Type, fk.ReflectField().Type)
						return errors.WrapDetf(ErrModelDefinition, "foreign key type doesn't match the primary field type of the root model")
					}

					relationship.mtmRelatedForeignKey = m2mFK
				} else {
					relationship.setKind(RelHasMany)
					if foreignKey == "" {
						// the foreign key for any of the slice relationships should be
						// model's that contains the relationship name concatenated with the 'ID'.
						foreignKey = relField.Name() + "ID"
					}

					modelWithFK := relationship.mStruct
					// get the name from the NamingConvention.
					foreignKeyName := m.Options.Namer.Namer(foreignKey)

					// check if given FK exists in the model's definitions.
					fk, ok := modelWithFK.ForeignKey(foreignKeyName)
					if !ok {
						foreignKey = model.modelType.Name() + "ID"
						fk, ok = modelWithFK.ForeignKey(m.Options.Namer.Namer(foreignKey))
						if !ok {
							log.Errorf("Foreign key: '%s' not found within Model: '%s'", foreignKeyName, modelWithFK.Type().Name())
							return errors.WrapDetf(ErrModelDefinition, "Foreign key: '%s' not found for the relationship: '%s'. Model: '%s'", foreignKeyName, relField.Name(), model.Type().Name())
						}
					}

					fkType := fk.ReflectField().Type
					if fkType.Kind() == reflect.Ptr {
						fkType = fkType.Elem()
					}
					// the primary field type of the model should match current's model type.
					if model.Primary().ReflectField().Type != fkType {
						log.Debugf("the foreign key in model: %v for the to-many relation: '%s' doesn't match the primary field type. Wanted: %v, Is: %v", modelWithFK.Type().Name(), relField.Name(), model.Type().Name(), model.Primary().ReflectField().Type, fkType)
						return errors.WrapDetf(ErrModelDefinition, "foreign key type doesn't match the primary field type of the root model")
					}
					relationship.setForeignKey(fk)
				}
			case reflect.Ptr, reflect.Struct:
				// Check if it is belongs_to or has_one relationship.
				if foreignKey == "" {
					// If foreign key has no given name the default value
					// is the relationship field name concatenated with 'ID'.
					foreignKey = relField.ReflectField().Name + "ID"
				}
				// Use the namer func to get the field's name
				foreignKeyName := m.Options.Namer.Namer(foreignKey)

				// Search for the foreign key within the given model
				fk, ok := model.ForeignKey(foreignKeyName)
				if ok {
					relationship.setKind(RelBelongsTo)
					relationship.foreignKey = fk
					continue
				}

				// If the foreign key is not found it must be a has one model or an invalid field name was provided.
				if !taggedForeign {
					// check if the model might have a name of belong's to
					modelsForeign := relField.relationship.mStruct.Type().Name() + "ID"
					// check if the foreign key would be the name of the related structure
					relatedTypeName := m.Options.Namer.Namer(modelsForeign)
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

				fk, ok = model.findUntypedInvalidAttribute(foreignKeyName)
				if ok {
					relationship.setKind(RelBelongsTo)
					relationship.foreignKey = fk
					continue
				}

				// if none of the foreign were found within the 'model', try to find it within
				// related model. It would be then a 'HasOne' relationship
				fk, ok = relationship.mStruct.ForeignKey(foreignKeyName)
				if ok {
					relationship.kind = RelHasOne
					// set the foreign key for the given relationship
					relationship.foreignKey = fk
					continue
				}

				fk, ok = relationship.mStruct.findUntypedInvalidAttribute(foreignKeyName)
				if ok {
					relationship.kind = RelHasOne
					relationship.foreignKey = fk
					continue
				}

				modelsForeign := m.Options.Namer.Namer(relField.mStruct.Type().Name() + "ID")

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

				// Provided invalid foreign field name.
				return errors.WrapDetf(ErrModelDefinition, "foreign key: '%s' not found for the relationship: '%s'. Model: '%s'", foreignKeyName, relField.Name(), model.Type().Name())
			}
		}

		for _, relField := range model.RelationFields() {
			if relField.relationship.isMany2Many() {
				continue
			}

			// check if the foreign key match related primary field type
			var match bool
			foreignKeyType := relField.relationship.foreignKey.ReflectField().Type
			switch relField.relationship.kind {
			case RelBelongsTo:
				if relField.relationship.mStruct.Primary().ReflectField().Type == foreignKeyType ||
					(foreignKeyType.Kind() == reflect.Ptr && foreignKeyType.Elem() == relField.relationship.mStruct.Primary().ReflectField().Type) {
					match = true
				}
			case RelHasMany, RelHasOne:
				if relField.mStruct.Primary().ReflectField().Type == foreignKeyType ||
					(foreignKeyType.Kind() == reflect.Ptr && foreignKeyType.Elem() == relField.mStruct.Primary().ReflectField().Type) {
					match = true
				}
			}

			if !match {
				log.Errorf("the foreign key in model: %v for the belongs-to relation: %s with model: %s is of invalid type. Wanted: %v, Is: %v", model,
					relField.relationship.foreignKey.mStruct.Type().Name(),
					relField.relationship.modelType.Name(),
					relField.relationship.RelatedModelStruct().Primary().ReflectField().Type,
					foreignKeyType)
				return errors.WrapDetf(ErrModelDefinition, "foreign key: '%s' doesn't match model's: '%s' primary key type", relField.relationship.foreignKey.Name(), relField.relationship.mStruct.Type().Name())
			}
		}
	}

	for _, model := range m.collections {
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
		if strings.EqualFold(field.Name(), "id") {
			err = model.setPrimaryField(field)
			if err != nil {
				return err
			}
			continue
		}

		// if strings.ToLower(field.Name())
		_, ok := reflect.New(field.BaseType()).Interface().(Model)
		if ok {
			// set them as the relationships
			err = model.setRelationshipField(field)
			if err != nil {
				return err
			}
			continue
		}

		if strings.HasSuffix(field.Name(), "ID") {
			var isForeignKey bool
			for _, other := range m.collections {
				if strings.HasPrefix(field.Name(), other.Type().Name()) {
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
func buildModelStruct(model interface{}, namer NamingConvention) (modelStruct *ModelStruct, err error) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		err = errors.WrapDetf(ErrModelDefinition, `provided model in invalid format. The model must be of struct or ptr type, but is: %v`, modelType)
		return nil, err
	}

	modelValue := reflect.New(modelType).Elem()

	var collection string
	// Check if the model implements Collectioner interface
	collectionNamer, ok := model.(Collectioner)
	if ok {
		collection = collectionNamer.NeuronCollectionName()
	} else {
		collection = namer.Namer(inflection.Plural(modelType.Name()))
	}

	modelStruct = newModelStruct(modelType, collection)
	modelStruct.namer = namer
	// map fields
	if err := modelStruct.mapFields(modelType, modelValue, nil); err != nil {
		return nil, err
	}
	return modelStruct, nil
}

func getNestedStruct(t reflect.Type, sFielder StructFielder, namerFunc NamingConvention) (*NestedStruct, error) {
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
				marshalField.Tag = `json:"-"`
				marshalFields = append(marshalFields, marshalField)
				continue
			}

			// get the tag values
			tagValues := nestedField.structField.TagValues(tag)
			for tKey, tValue := range tagValues {
				switch tKey {
				case AnnotationName:
					nestedField.structField.neuronName = tValue[0]
				case AnnotationFieldType:
					if tValue[0] != AnnotationNestedField {
						log.Debugf("Invalid annotation.NestedField value: '%s' for field: %s", tValue[0], nestedField.structField.Name())
						return nil, errors.WrapDetf(ErrModelDefinition, "provided field type: '%s' is not allowed for the nested struct field: '%s'", nestedField.structField.reflectField.Type, nestedField.structField.Name())
					}
				case AnnotationFlags:
					if err := nestedField.structField.setFlags(tValue...); err != nil {
						return nil, err
					}
				}
			}
		}

		if nestedField.structField.NeuronName() == "" {
			nestedField.structField.neuronName = namerFunc.Namer(nField.Name)
		}

		switch nestedField.structField.NeuronName() {
		case "relationships", "links":
			return nil, errors.WrapDetf(ErrModelDefinition, "nested field within: '%s' field in the model: '%s' has forbidden Neuron name: '%s'",
				nestedStruct.Attr().Name(),
				nestedStruct.Attr().Struct().Type().Name(),
				nestedField.structField.NeuronName(),
			)
		}

		if _, ok = NestedStructSubField(nestedStruct, nestedField.structField.NeuronName()); ok {
			return nil, errors.WrapDetf(ErrModelDefinition, "nestedStruct: %v already has one nestedField: '%s'. The fields must be uniquely named", nestedStruct.Type().Name(), nestedField.structField.Name())
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
					// disallow nested map, arrays, maps in ptr type slices
					return nil, errors.WrapDetf(ErrModelDefinition, "structField: '%s' nested type is invalid. The model doesn't allow one of slices to ptr of slices or map", nestedField.structField.Name())
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
				return nil, errors.WrapDetf(ErrModelDefinition, "nested field can't be a slice of pointer to slices|map|arrays. NestedField: '%s' within NestedStruct:'%s'", nestedField.structField.Name(), nestedStruct.modelType.Name())
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
