package models

import (
	"fmt"
	"github.com/jinzhu/inflection"
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/flags"
	"github.com/neuronlabs/neuron/internal/namer"
	"github.com/neuronlabs/neuron/log"
	"github.com/pkg/errors"
	"reflect"
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

	// Flags contains the config flags for given schema
	Flags *flags.Container

	// NamerFunc is the function required for naming convenction
	NamerFunc namer.Namer
}

// NewModelSchemas create new ModelSchemas based on the config
func NewModelSchemas(
	namerFunc namer.Namer,
	c *config.Controller,
	flgs *flags.Container,
) (*ModelSchemas, error) {
	return newModelSchemas(namerFunc,
		c.ModelSchemas,
		c.DefaultSchema,
		c.DefaultRepositoryName,
		flgs,
	)
}

func newModelSchemas(
	namerFunc namer.Namer,
	cfg map[string]*config.Schema,
	defaultSchema string,
	defaultRepoName string,
	flgs *flags.Container,
) (*ModelSchemas, error) {
	log.Debugf("Creating New ModelSchemas...")
	ms := &ModelSchemas{
		Flags:           flgs,
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
func (m *ModelSchemas) RegisterModels(
	models ...interface{},
) error {

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
		mStruct, err := BuildModelStruct(model, m.NamerFunc, m.Flags)
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

func (m *ModelSchemas) setModelRelationships(model *ModelStruct) (err error) {
	schema := m.schemas[model.SchemaName()]

	for _, relField := range model.RelationshipFields() {
		relType := FieldsRelatedModelType(relField)

		val := schema.models.Get(relType)
		if val == nil {
			err = fmt.Errorf("Model: %v, not precalculated but is used in relationships for: %v field in %v model.", relType, relField.FieldName(), model.Type().Name())
			return err
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
				err = fmt.Errorf("Join Model not found for model: '%s' relationship: '%s'", model.Type().Name(), relField.Name())
				return
			}

			rel.joinModel.isJoin = true

			// check if the name was preset before if not set it to the model's type name + ID
			var isDefault bool
			if rel.backReferenceForeignKeyName == "" {
				rel.backReferenceForeignKeyName = model.Type().Name() + "ID"
				isDefault = true
			}

			fk, ok := rel.joinModel.ForeignKey(rel.backReferenceForeignKeyName)
			if !ok {
				if isDefault {
					err = fmt.Errorf("Backreference foreign key with default name: '%s' not found within the join model: '%s' ", rel.backReferenceForeignKeyName, rel.joinModel.Type().Name())
				} else {
					err = fmt.Errorf("Backreference foreign key: '%s' not found within the join model: '%s' ", rel.backReferenceForeignKeyName, rel.joinModel.Type().Name())
				}
				return
			}

			rel.backReferenceForeignKey = fk
		}
	}

	return
}

func (m *ModelSchemas) setRelationships() (err error) {
	for _, schema := range m.schemas {

		// iterate over all models from schema
		for _, model := range schema.models.Models() {

			for _, relField := range model.RelationshipFields() {

				// relationship gets the relationship between the fields
				relationship := relField.Relationship()

				// get structfield jsonapi tags
				tags := relField.TagValues(relField.ReflectField().Tag.Get(internal.AnnotationNeuron))

				// get proper foreign key field name
				fkeyFieldName := tags.Get(internal.AnnotationForeignKey)

				log.Debugf("Relationship field: %s, foreign key name: %s", relField.Name(), fkeyFieldName)
				// check field type

				switch relField.ReflectField().Type.Kind() {
				case reflect.Slice:

					// get the foreign key name
					if fkeyFieldName == "" {
						fkeyFieldName = relField.relationship.modelType.Name() + "ID"
					}

					var modelWithFK *ModelStruct
					if relationship.IsManyToMany() {
						modelWithFK = relationship.joinModel
					} else {
						modelWithFK = relField.relationship.mStruct
					}

					fkeyName := m.NamerFunc(fkeyFieldName)
					fk, ok := modelWithFK.ForeignKey(fkeyName)
					if !ok {
						err = errors.Errorf("Foreign key not found for the relationship: '%s'. Model: '%s'", relField.Name(), model.Type().Name())
						return
					}

					if model.PrimaryField().ReflectField().Type != fk.ReflectField().Type {

						err = errors.Errorf("The foreign key in model: %v for the has-many relation: %s within model: %s is of invalid type. Wanted: %v, Is: %v",
							FieldsRelatedModelType(fk).Name(),
							relField.Name(),
							model.Type().Name(),
							model.PrimaryField().ReflectField().Type,
							fk.ReflectField().Type,
						)
						return
					}
					relationship.SetForeignKey(fk)
					if relationship.Kind() != RelMany2Many {
						relationship.kind = RelHasMany
					}

				case reflect.Ptr, reflect.Struct:

					// check if it is belongs_to or has_one relationship
					// at first search for foreign key as
					if fkeyFieldName == "" {
						// if foreign key has no predefined name
						// the default value is the relationship field name + ID
						fkeyFieldName = relField.ReflectField().Name + "ID"
					}

					// get the foreign key name
					fkeyName := m.NamerFunc(fkeyFieldName)

					// search for the foreign key within the given model
					fk, ok := model.ForeignKey(fkeyName)
					if !ok {

						// if the foreign key is not found it must be a has one model or an invalid field name was provided
						relationship.SetKind(RelHasOne)

						fk, ok = relationship.mStruct.ForeignKey(fkeyName)
						if !ok {
							// provided invalid foreign field name
							err = errors.Errorf("Foreign key not found for the relationship: '%s'. Model: '%s'", relField.Name(), model.Type().Name())
							return
						}

						// then the check if the current model's primary field is of the same type as the foreign key
						if model.PrimaryField().ReflectField().Type != fk.ReflectField().Type {
							err = errors.Errorf("The foreign key in model: %v for the has-one relation: %s within model: %s is of invalid type. Wanted: %v, Is: %v",
								fk.Struct().Type().Name(),
								relField.Name(),
								model.Type().Name(),
								model.PrimaryField().ReflectField().Type,
								fk.ReflectField().Type)
							return
						}

					} else {
						// relationship is of BelongsTo kind

						// check if the related model struct primary field is of the same type as the given foreing key
						if relationship.mStruct.PrimaryField().ReflectField().Type != fk.ReflectField().Type {
							err = errors.Errorf("The foreign key in model: %v for the belongs-to relation: %s with model: %s is of invalid type. Wanted: %v, Is: %v", model,
								relField.Name(),
								FieldsRelatedModelType(relField).Name(),
								FieldsRelatedModelStruct(relField).PrimaryField().ReflectField().Type,
								fk.ReflectField().Type,
							)
							return
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
		return nil, internal.ErrModelNotMapped
	}

	mStruct := schema.models.Get(t)
	if mStruct == nil {
		return nil, internal.ErrModelNotMapped
	}

	return mStruct, nil
}

func (m *ModelSchemas) getModelStruct(model interface{}) (*ModelStruct, error) {
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
		return nil, internal.ErrModelNotMapped
	}
	return schema, nil
}
