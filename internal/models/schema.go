package models

import (
	"reflect"

	"github.com/jinzhu/inflection"

	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/errors/class"
	"github.com/neuronlabs/neuron/log"

	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/namer"
)

// Schema is a container for the given
type Schema struct {
	config *config.Schema

	// models contains model definition per single schema
	models *ModelMap

	// Name is the schema name for given models
	// - i.e. public models would have the name 'public'
	Name string
}

// Model returns model for given type within the schema
func (s *Schema) Model(t reflect.Type) *ModelStruct {
	return s.models.Get(t)
}

// Models returns all the models saved in the given schema
func (s *Schema) Models() []*ModelStruct {
	return s.models.Models()
}

// ModelByCollection returns ModelStruct on the base of provided collection name
func (s *Schema) ModelByCollection(collection string) *ModelStruct {
	return s.models.GetByCollection(collection)
}

// ModelByName gets the model by it's struct name
func (s *Schema) ModelByName(name string) *ModelStruct {
	for _, model := range s.models.models {
		if model.Type().Name() == name {
			return model
		}
	}
	return nil
}

// Config gets the config.Schema
func (s *Schema) Config() *config.Schema {
	return s.config
}

// ModelSchemas is a struct containing all the schemas mapped with it's model's and names
type ModelSchemas struct {
	cfg map[string]*config.Schema

	schemas      map[string]*Schema
	schemaByType map[reflect.Type]*Schema

	defaultSchema   *Schema
	defaultRepoName string

	// NamerFunc is the function required for naming convenction
	NamerFunc namer.Namer
}

// NewModelSchemas create new ModelSchemas based on the config
func NewModelSchemas(
	namerFunc namer.Namer,
	c *config.Controller,
) (*ModelSchemas, error) {
	return newModelSchemas(namerFunc,
		c.ModelSchemas,
		c.DefaultSchema,
		c.DefaultRepositoryName,
	)
}

func newModelSchemas(
	namerFunc namer.Namer,
	cfg map[string]*config.Schema,
	defaultSchema string,
	defaultRepoName string,
) (*ModelSchemas, error) {
	log.Debugf("Creating New ModelSchemas...")
	ms := &ModelSchemas{
		NamerFunc:       namerFunc,
		cfg:             cfg,
		defaultRepoName: defaultRepoName,
	}

	ms.defaultSchema = &Schema{Name: defaultSchema, models: NewModelMap()}
	ms.schemas = make(map[string]*Schema)
	ms.schemas[defaultSchema] = ms.defaultSchema

	// set the schema's configs
	for name, schemaCfg := range cfg {
		if name == defaultSchema {
			ms.defaultSchema.config = schemaCfg
		} else {
			ms.schemas[name] = &Schema{
				config: schemaCfg,
				Name:   name,
				models: NewModelMap(),
			}
		}
		log.Debugf("Schema %s created with config", name)
	}

	return ms, nil
}

// ComputeNestedIncludedCount computes the limits for the nested included count for each model in each schema
func (m *ModelSchemas) ComputeNestedIncludedCount(limit int) {
	for _, schema := range m.schemas {
		for _, model := range schema.Models() {
			model.initComputeThisIncludedCount()
		}
	}

	for _, schema := range m.schemas {
		for _, model := range schema.Models() {
			model.computeNestedIncludedCount(limit)
		}
	}
}

// DefaultSchema returns default schema for give models
func (m *ModelSchemas) DefaultSchema() *Schema {
	return m.defaultSchema
}

// Schema returns schema on the base of schemaName
func (m *ModelSchemas) Schema(schemaName string) (*Schema, bool) {
	s, ok := m.schemas[schemaName]
	return s, ok
}

// Schemas returns all registered schemas
func (m *ModelSchemas) Schemas() []*Schema {
	var schemas []*Schema

	for _, schema := range m.schemas {
		schemas = append(schemas, schema)
	}

	return schemas
}

// RegisterModels registers the model within the schemas container
func (m *ModelSchemas) RegisterModels(models ...interface{}) error {
	// iterate over models and register one by one
	for _, model := range models {
		var schema string

		// set model's schema
		if schemaNamer, ok := model.(SchemaNamer); ok {
			schema = schemaNamer.SchemaName()
		} else {
			schema = m.defaultSchema.Name
		}

		// check if schema is already created
		s, ok := m.schemas[schema]
		if !ok {
			log.Debugf("Schema: %s not found for the model. Creating new schema.", schema)
			s = &Schema{
				Name:   schema,
				models: NewModelMap(),
			}

			m.schemas[schema] = s
		}

		// build the model's structure and set into schema's model map
		mStruct, err := buildModelStruct(model, m.NamerFunc)
		if err != nil {
			return err
		}

		mStruct.SetSchemaName(schema)

		if err := s.models.Set(mStruct); err != nil {
			continue
		}

		var modelConfig *config.ModelConfig

		if s.config == nil {
			s.config = &config.Schema{
				Name:   schema,
				Models: map[string]*config.ModelConfig{},
				Local:  true,
			}
		}

		if s.config.Models == nil {
			s.config.Models = map[string]*config.ModelConfig{}
		}

		modelConfig, ok = s.config.Models[mStruct.collectionType]
		if !ok {
			modelConfig = &config.ModelConfig{}

			s.config.Models[mStruct.collectionType] = modelConfig
		}
		modelConfig.Collection = mStruct.Collection()

		log.Debugf("Getting model config from schema: '%s'", schema)
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

	for _, schema := range m.schemas {
		for _, model := range schema.models.Models() {
			if err := m.setModelRelationships(model); err != nil {
				return err
			}

			if err := model.initCheckFieldTypes(); err != nil {
				return err
			}
			model.initComputeSortedFields()

			schema.models.SetByCollection(model)
		}
	}

	err := m.setRelationships()
	if err != nil {
		return err
	}

	return nil
}

// RegisterSchemaModels registers models for provided schema
func (m *ModelSchemas) RegisterSchemaModels(schemaName string, models ...interface{}) error {
	/**

	TO DO:

	- Get Schema from name
	- register models just for this single schema


	*/
	return nil
}

// RegisterModelsRecursively registers the models and it's related models into the model schemas
func (m *ModelSchemas) RegisterModelsRecursively(models ...interface{}) error {
	/**

	TO DO:

	-iterate over models
	- register models
	- if the related models are not registered, recursively check and register related models

	*/
	return nil
}

func (m *ModelSchemas) setModelRelationships(model *ModelStruct) error {
	schema := m.schemas[model.SchemaName()]

	for _, relField := range model.RelationshipFields() {
		relType := FieldsRelatedModelType(relField)

		val := schema.models.Get(relType)
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

			for _, other := range schema.models.models {
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

func (m *ModelSchemas) setRelationships() error {
	for _, schema := range m.schemas {
		// iterate over all models from schema
		for _, model := range schema.models.Models() {
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
	}

	for _, schema := range m.schemas {
	modelLoop:
		for _, model := range schema.models.Models() {
			for _, relField := range model.relationships {
				if relField.Relationship().kind != RelBelongsTo {
					model.StoreSet(hasForeignRelationships, struct{}{})
					break modelLoop
				}
			}
		}
	}
	return nil
}

// GetModelStruct gets the model from the model schemas
func (m *ModelSchemas) GetModelStruct(model interface{}) (*ModelStruct, error) {
	return m.getModelStruct(model)
}

// ModelByType returns the model on the base of the provided model type
func (m *ModelSchemas) ModelByType(t reflect.Type) (*ModelStruct, error) {
	return m.getByType(t)
}

// SchemaByType returns schema by the type provided in the arguments
func (m *ModelSchemas) SchemaByType(t reflect.Type) (*Schema, error) {
	return m.getSchemaByType(t)
}

func (m *ModelSchemas) getByType(t reflect.Type) (*ModelStruct, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}

	var schemaName string
	schemaNamer, ok := reflect.New(t).Interface().(SchemaNamer)
	if ok {
		schemaName = schemaNamer.SchemaName()
		log.Debugf("Schema Namer: %T", schemaNamer)
	} else {
		schemaName = m.defaultSchema.Name
	}

	schema, ok := m.schemas[schemaName]
	if !ok {
		return nil, errors.Newf(class.ModelSchemaNotFound, "schema: '%s' not found", schemaName)
	}

	mStruct := schema.models.Get(t)
	if mStruct == nil {
		return nil, errors.Newf(class.ModelNotMappedInSchema, "model: '%s' is not found within schema: '%s'", t.Name(), schemaName)
	}

	return mStruct, nil
}

func (m *ModelSchemas) getModelStruct(model interface{}) (*ModelStruct, error) {
	mStruct, ok := model.(*ModelStruct)
	if ok {
		return mStruct, nil
	}
	t := reflect.TypeOf(model)
	return m.getByType(t)
}

func (m *ModelSchemas) getSchemaByType(t reflect.Type) (*Schema, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}

	var schemaName string
	schemaNamer, ok := reflect.New(t).Interface().(SchemaNamer)
	if ok {
		schemaName = schemaNamer.SchemaName()
	} else {
		schemaName = m.defaultSchema.Name
	}

	schema, ok := m.schemas[schemaName]
	if !ok {
		return nil, errors.Newf(class.ModelSchemaNotFound, "model schema: '%s' not found", schemaName)
	}
	return schema, nil
}
