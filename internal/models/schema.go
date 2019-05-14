package models

import (
	"fmt"
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

	// nestedIncludeLimit is the config used for mapping the models
	NestedIncludeLimit int
}

// NewModelSchemas create new ModelSchemas based on the config
func NewModelSchemas(
	namerFunc namer.Namer,
	c *config.ControllerConfig,
	flgs *flags.Container,
) (*ModelSchemas, error) {
	return newModelSchemas(namerFunc, c.Builder.IncludeNestedLimit, c.ModelSchemas,
		c.DefaultSchema, c.DefaultRepository, flgs)
}

func newModelSchemas(
	namerFunc namer.Namer,
	nestedIncludeLimit int,
	cfg map[string]*config.Schema,
	defaultSchema string,
	defaultRepoName string,
	flgs *flags.Container,
) (*ModelSchemas, error) {
	log.Debugf("Creating New ModelSchemas...")
	ms := &ModelSchemas{
		NestedIncludeLimit: nestedIncludeLimit,
		Flags:              flgs,
		NamerFunc:          namerFunc,
		cfg:                cfg,
		defaultRepoName:    defaultRepoName,
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

		if s.config != nil {
			log.Debugf("Getting model config from schema: '%s'", schema)
			modelConfig, ok := s.config.Models[mStruct.collectionType]
			if ok {
				if err := mStruct.SetConfig(modelConfig); err != nil {
					log.Errorf("Setting config for model: '%s' failed.", mStruct.Collection())
					return err
				}
			}
		}

		if mStruct.repositoryName == "" {
			// if the model implements repository Name
			repositoryName, ok := model.(namer.RepositoryNamer)
			if ok {
				mStruct.repositoryName = repositoryName.RepositoryName()
			} else {
				mStruct.repositoryName = m.defaultRepoName
			}
		}

	}

	for _, schema := range m.schemas {
		for _, model := range schema.models.Models() {
			if err := m.setModelRelationships(model); err != nil {
				return err
			}

			if err := InitCheckFieldTypes(model); err != nil {
				return err
			}
			InitComputeSortedFields(model)

			InitComputeThisIncludedCount(model)

			schema.models.SetByCollection(model)
		}
	}

	for _, schema := range m.schemas {
		for _, model := range schema.models.Models() {
			model.InitComputeNestedIncludedCount(m.NestedIncludeLimit)
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

	for _, rel := range model.RelationshipFields() {
		relType := FieldsRelatedModelType(rel)
		val := schema.models.Get(relType)
		if val == nil {
			err = fmt.Errorf("Model: %v, not precalculated but is used in relationships for: %v field in %v model.", relType, rel.FieldName(), model.Type().Name())
			return err
		}
		rel.SetRelatedModel(val)
	}

	return
}

func (m *ModelSchemas) setRelationships() error {
	for _, schema := range m.schemas {
		for _, model := range schema.models.Models() {
			for _, relField := range model.RelationshipFields() {

				relationship := relField.Relationship()

				// get structfield jsonapi tags
				tags, err := relField.TagValues(relField.ReflectField().Tag.Get(internal.AnnotationNeuron))
				if err != nil {
					return err
				}

				// get proper foreign key field name
				fkeyFieldName := tags.Get(internal.AnnotationForeignKey)

				log.Debugf("Relationship field: %s, foreign key name: %s", relField.Name(), fkeyFieldName)
				// check field type
				switch relField.ReflectField().Type.Kind() {
				case reflect.Slice:
					// has many by default
					if relationship.IsManyToMany() {
						// if relationship.Sync != nil && !(*relationship.Sync) {
						// 	continue
						// }
						if bf := relationship.BackrefernceFieldName(); bf != "" {
							bf = m.NamerFunc(bf)
							backReferenced, ok := relField.Relationship().Struct().RelationshipField(bf)
							if !ok {
								err = errors.Errorf("The backreference collection named: '%s' is invalid. Model: %s, Sfield: '%s'", bf, model.Type().Name(), relField.ReflectField().Name)
								return err
							}

							mustBeType := reflect.SliceOf(reflect.New(model.Type()).Type())

							if backReferenced.ReflectField().Type != mustBeType {
								err = errors.Errorf("The backreference field for relation: %v within model: %v   is of invalid type. Wanted: %v. Is: %v", relField.Name(), model.Type().Name(), mustBeType, backReferenced.ReflectField().Type)
								return err
							}

							relationship.SetBackreferenceField(backReferenced)

						}
						continue
					}

					// HasMany
					relationship.SetKind(RelHasMany)

					if fkeyFieldName == "" {
						fkeyFieldName = model.Type().Name() + "ID"
					}

					fkeyName := m.NamerFunc(fkeyFieldName)
					fk, ok := FieldsRelatedModelStruct(relField).ForeignKey(fkeyName)
					if !ok {
						return errors.Errorf("Foreign key not found for the relationship: '%s'. Model: '%s'", relField.Name(), model.Type().Name())
					}

					if model.PrimaryField().ReflectField().Type != fk.ReflectField().Type {

						return errors.Errorf("The foreign key in model: %v for the has-many relation: %s within model: %s is of invalid type. Wanted: %v, Is: %v",
							FieldsRelatedModelType(fk).Name(),
							relField.Name(),
							model.Type().Name(),
							model.PrimaryField().ReflectField().Type,
							fk.ReflectField().Type,
						)
					}
					relationship.SetForeignKey(fk)

					if relationship.Sync() != nil && !(*relationship.Sync()) {
						// c.log().Debugf("Relationship: %s is non-synced.", relField.fieldName)
						continue
					}

					b := true

					relationship.SetSync(&b)

				case reflect.Ptr, reflect.Struct:
					// check if it is belongs_to or has_one relationship
					// at first search for foreign key as
					if fkeyFieldName == "" {
						fkeyFieldName = relField.ReflectField().Name + "ID"
					}
					fkeyName := m.NamerFunc(fkeyFieldName)
					nosync := (relationship.Sync() != nil && !*relationship.Sync())
					// c.log().Debugf("Model: %v Looking for foreignkey: %s", model.modelType.Name(), fkeyName)
					fk, ok := model.ForeignKey(fkeyName)
					if !ok {
						// c.log().Debugf("Not found within root model for relation: %s, foreign: %s", relField.fieldName, fkeyFieldName)
						relationship.SetKind(RelHasOne)
						fk, ok = FieldsRelatedModelStruct(relField).ForeignKey(fkeyName)
						if !ok {
							return errors.Errorf("Foreign key not found for the relationship: '%s'. Model: '%s'", relField.Name(), model.Type().Name())
						}

						if model.PrimaryField().ReflectField().Type != fk.ReflectField().Type {
							return errors.Errorf("The foreign key in model: %v for the has-one relation: %s within model: %s is of invalid type. Wanted: %v, Is: %v",
								fk.Struct().Type().Name(),
								relField.Name(),
								model.Type().Name(),
								model.PrimaryField().ReflectField().Type,
								fk.ReflectField().Type)
						}
						sync := !nosync
						relationship.SetSync(&sync)
						// c.log().Debugf("Found within related model: %v field: %v", FieldsRelatedModelType(relField).Name(), fk.Name())

					} else {
						// c.log().Debugf("found for: %s", relField.fieldName)
						if FieldsRelatedModelStruct(relField).PrimaryField().ReflectField().Type != fk.ReflectField().Type {
							return errors.Errorf("The foreign key in model: %v for the belongs-to relation: %s with model: %s is of invalid type. Wanted: %v, Is: %v", model,
								relField.Name(),
								FieldsRelatedModelType(relField).Name(),
								FieldsRelatedModelStruct(relField).PrimaryField().ReflectField().Type,
								fk.ReflectField().Type,
							)
						}
						relationship.SetKind(RelBelongsTo)
					}
					relationship.SetForeignKey(fk)
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
		return nil, internal.IErrModelNotMapped
	}

	mStruct := schema.models.Get(t)
	if mStruct == nil {
		return nil, internal.IErrModelNotMapped
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
		return nil, internal.IErrModelNotMapped
	}
	return schema, nil
}
