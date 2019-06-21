package controller

import (
	"strings"

	"gopkg.in/go-playground/validator.v9"

	"github.com/neuronlabs/uni-logger"

	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/errors/class"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"

	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/namer"
	"github.com/neuronlabs/neuron/internal/query/scope"
)

var (
	validate          = validator.New()
	defaultController *Controller
)

// Controller is the data structure that is responsible for controlling all the models
// within single API
type Controller struct {
	// Config is the configuration struct for the controller
	Config *config.Controller

	// Namer defines the function strategy how the model's and it's fields are being named
	NamerFunc namer.Namer

	// StrictUnmarshalMode if set to true, the incoming data cannot contain
	// any unknown fields
	StrictUnmarshalMode bool

	// // queryBuilderis the controllers query builder
	// queryBuilder *query.Builder

	processor scope.Processor

	// schemas is a mapping for the model schemas
	schemas *models.ModelSchemas

	// Validators
	// CreateValidator is used as a validator for the Create processes
	CreateValidator *validator.Validate

	//PatchValidator is used as a validator for the Patch processes
	PatchValidator *validator.Validate
}

// New Creates raw *jsonapi.Controller with no limits and links.
func New(cfg *config.Controller, logger unilogger.LeveledLogger) (*Controller, error) {
	if logger != nil {
		log.SetLogger(logger)
	} else {
		log.Default()
	}

	log.Debugf("New Controller creating...")
	c, err := newController(cfg)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// SetDefault sets the default controller
func SetDefault(c *Controller) {
	defaultController = c
}

// Default creates new *jsonapi.Controller with preset limits:
// Controller has also set the FlagUseLinks flag to true.
func Default() *Controller {
	if defaultController == nil {

		c, err := newController(DefaultConfig)
		if err != nil {
			panic(err)
		}

		defaultController = c
	}

	return defaultController
}

func newController(cfg *config.Controller) (*Controller, error) {
	var err error
	c := &Controller{
		CreateValidator: validator.New(),
		PatchValidator:  validator.New(),
	}

	log.Debugf("Creating Controller with config: %+v", cfg)

	if err = c.setConfig(cfg); err != nil {
		return nil, err
	}

	// create model schemas
	c.schemas, err = models.NewModelSchemas(
		c.NamerFunc,
		c.Config,
	)
	if err != nil {
		return nil, err
	}

	defaultRepository := c.Config.DefaultRepository
	if defaultRepository == nil {
		return nil, errors.New(class.ModelSchemaNotFound, "default repository not found")
	}

	// set default factory
	if factory := repository.GetFactory(defaultRepository.DriverName); factory == nil {
		return nil, errors.Newf(class.RepositoryNotFound, "repository Factory not found for the repository: %s", defaultRepository.DriverName)
	}

	return c, nil
}

// SetLogger sets the logger for the controller operations
func (c *Controller) SetLogger(logger unilogger.LeveledLogger) {
	log.SetLogger(logger)
}

// ModelStruct gets the model struct  mapping.
// Implements repository.ModelStructer
func (c *Controller) ModelStruct(model interface{}) (*mapping.ModelStruct, error) {
	m, err := c.getModelStruct(model)
	if err != nil {
		return nil, err
	}
	return (*mapping.ModelStruct)(m), nil
}

// MustGetModelStruct gets (concurrently safe) the model struct from the cached model Map
// panics if the model does not exists in the map.
func (c *Controller) MustGetModelStruct(model interface{}) *models.ModelStruct {
	mStruct, err := c.getModelStruct(model)
	if err != nil {
		panic(err)
	}
	return mStruct
}

// RegisterModels precomputes provided models, making it easy to check
// models relationships and  attributes.
func (c *Controller) RegisterModels(models ...interface{}) error {
	if err := c.schemas.RegisterModels(models...); err != nil {
		return err
	}

	for _, schema := range c.schemas.Schemas() {
		if err := c.Config.MapRepositories(schema.Config()); err != nil {
			log.Debugf("Mapping repositories failed for schema: %s", schema.Name)
			return err
		}

		for _, mStruct := range schema.Models() {
			if _, err := repository.GetRepository(c, (*mapping.ModelStruct)(mStruct)); err != nil {
				log.Errorf("Mapping model: %v to repository failed.", mStruct.Type().Name())
				return err
			}
		}
	}
	return nil
}

// RegisterSchemaModels registers the model for the provided schema
func (c *Controller) RegisterSchemaModels(schemaName string, models ...interface{}) error {
	return c.ModelSchemas().RegisterSchemaModels(schemaName, models...)
}

// RegisterModelRecursively registers provided models and it's realtionship fields recursively
func (c *Controller) RegisterModelRecursively(models ...interface{}) error {
	return c.ModelSchemas().RegisterModelsRecursively(models...)
}

// GetModelStruct returns the ModelStruct for provided model
// Returns error if provided model does not exists in the PrecomputedMap
func (c *Controller) GetModelStruct(model interface{}) (*models.ModelStruct, error) {
	return c.getModelStruct(model)
}

func (c *Controller) getModelStruct(model interface{}) (*models.ModelStruct, error) {
	if model == nil {
		return nil, errors.New(class.ModelValueNil, "provided nil model value")
	}

	mStruct, err := c.schemas.GetModelStruct(model)
	if err != nil {
		return nil, err
	}

	return mStruct, nil
}

// setConfig sets and validates provided config
func (c *Controller) setConfig(cfg *config.Controller) error {
	if cfg == nil {
		return errors.New(class.ConfigValueNil, "provided nil config value")
	}

	// set level debug
	if cfg.Debug {
		log.SetLevel(log.LDEBUG)
	}

	cfg.NamingConvention = strings.ToLower(cfg.NamingConvention)

	if err := validate.Struct(cfg); err != nil {
		return errors.New(class.ConfigValueInvalid, "validating config failed")
	}

	if err := cfg.Processor.Validate(); err != nil {
		return err
	}

	c.Config = cfg

	// set naming convention
	switch cfg.NamingConvention {
	case "kebab":
		c.NamerFunc = namer.NamingKebab
	case "camel":
		c.NamerFunc = namer.NamingCamel
	case "lowercamel":
		c.NamerFunc = namer.NamingLowerCamel
	case "snake":
		c.NamerFunc = namer.NamingSnake
	}
	log.Debugf("Naming Convention used in schemas: %s", cfg.NamingConvention)

	if cfg.DefaultSchema == "" {
		cfg.DefaultSchema = "api"
	}

	if err := cfg.SetDefaultRepository(); err != nil {
		return err
	}

	if cfg.CreateValidatorAlias == "" {
		cfg.CreateValidatorAlias = "create"
	}

	c.CreateValidator.SetTagName(cfg.CreateValidatorAlias)

	if cfg.PatchValidatorAlias == "" {
		cfg.PatchValidatorAlias = "patch"
	}
	c.PatchValidator.SetTagName(cfg.PatchValidatorAlias)

	return nil
}

// ModelSchemas returns model schemas for given controller
func (c *Controller) ModelSchemas() *models.ModelSchemas {
	return c.schemas
}

// Processor returns the query processor for the controller
func (c *Controller) Processor() scope.Processor {
	return c.processor
}

// SetProcessor sets the query processor for the controller
func (c *Controller) SetProcessor(p scope.Processor) {
	c.processor = p
}
