package controller

import (
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/i18n"
	"github.com/neuronlabs/neuron/internal/flags"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/query"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"

	aerrors "github.com/neuronlabs/neuron/errors"
	"github.com/pkg/errors"

	"github.com/kucjac/uni-logger"
	"github.com/neuronlabs/neuron/internal/namer"

	// "golang.org/x/text/language"
	// "golang.org/x/text/language/display"
	"gopkg.in/go-playground/validator.v9"
	// "net/http"
	// "net/url"

	// "strconv"
	"strings"
)

var (
	validate          *validator.Validate = validator.New()
	defaultController *Controller
)

// Controller is the data structure that is responsible for controlling all the models
// within single API
type Controller struct {
	// Config is the configuration struct for the controller
	Config *config.ControllerConfig

	// Namer defines the function strategy how the model's and it's fields are being named
	NamerFunc namer.Namer

	// Flags defines the controller config flags
	Flags *flags.Container

	// StrictUnmarshalMode if set to true, the incoming data cannot contain
	// any unknown fields
	StrictUnmarshalMode bool

	// queryBuilderis the controllers query builder
	queryBuilder *query.Builder

	// i18nSup defines the i18n support for the provided controller
	i18nSup *i18n.Support

	// schemas is a mapping for the model schemas
	schemas *models.ModelSchemas

	// dbErrMapper error manager for the repositories
	dbErrMapper *aerrors.ErrorMapper

	// Validators
	// CreateValidator is used as a validator for the Create processes
	CreateValidator *validator.Validate

	//PatchValidator is used as a validator for the Patch processes
	PatchValidator *validator.Validate
}

// New Creates raw *jsonapi.Controller with no limits and links.
func New(cfg *config.ControllerConfig, logger unilogger.LeveledLogger) (*Controller, error) {

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

func newController(cfg *config.ControllerConfig) (*Controller, error) {
	var (
		f   *flags.Container
		err error
	)
	if cfg.Flags == nil {
		f = flags.New()
	} else {
		f, err = cfg.Flags.Container()
		if err != nil {
			return nil, err
		}
	}

	c := &Controller{
		Flags:           f,
		CreateValidator: validator.New(),
		PatchValidator:  validator.New(),
	}

	log.Debugf("Creating Controller with config: %+v", cfg)

	if err = c.setConfig(cfg); err != nil {
		return nil, errors.Wrap(err, "setConfig failed.")
	}

	if cfg.I18n != nil {
		c.i18nSup, err = i18n.New(cfg.I18n)
		if err != nil {
			return nil, errors.Wrap(err, "i18n.New failed.")
		}
	}

	if c.Config.Builder == nil {
		return nil, errors.New("No builder within the controller config found")
	}

	// create model schemas
	c.schemas, err = models.NewModelSchemas(
		c.NamerFunc,
		c.Config,
		c.Flags,
	)
	if err != nil {
		return nil, err
	}

	defaultRepository := c.Config.DefaultRepository
	if defaultRepository == nil {
		return nil, errors.Errorf("Default repository: '%s' not found within repositories container", c.Config.DefaultRepository)
	}

	// set default factory
	if factory := repository.GetFactory(defaultRepository.DriverName); factory == nil {
		return nil, errors.Errorf("Repository Factory not found for the repository: %s ", defaultRepository.DriverName)
	}

	c.queryBuilder, err = query.NewBuilder(c.schemas, c.Config.Builder, c.i18nSup)
	if err != nil {
		return nil, errors.Wrap(err, "query.NewBuilder failed")
	}

	// create error manager
	c.dbErrMapper = aerrors.NewDBMapper()

	return c, nil
}

// DBErrorMapper gets the database error manager
func (c *Controller) DBErrorMapper() *aerrors.ErrorMapper {
	return c.dbErrMapper
}

// QueryBuilder returns the controller query builder
func (c *Controller) QueryBuilder() *query.Builder {
	return c.queryBuilder
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
		return nil, errors.New("Nil model provided.")
	}

	mStruct, err := c.schemas.GetModelStruct(model)
	if err != nil {
		return nil, err
	}

	return mStruct, nil
}

// setConfig sets and validates provided config
func (c *Controller) setConfig(cfg *config.ControllerConfig) error {
	if cfg == nil {
		return errors.New("Nil config provided")
	}

	// set level debug
	if cfg.Debug {
		log.SetLevel(log.LDEBUG)
	}

	cfg.NamingConvention = strings.ToLower(cfg.NamingConvention)

	if err := validate.Struct(cfg); err != nil {
		return errors.Wrap(err, "Validate config failed.")
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
