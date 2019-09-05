package controller

import (
	"context"
	"strings"
	"time"

	"gopkg.in/go-playground/validator.v9"

	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/uni-logger"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"
	"github.com/neuronlabs/neuron-core/namer"
	"github.com/neuronlabs/neuron-core/repository"
)

var (
	validate = validator.New()
)

// Controller is the structure that controls whole jsonapi behavior.
// It contains repositories, model definitions, query builders and it's own config.
type Controller struct {
	// Config is the configuration struct for the controller.
	Config *config.Controller
	// Namer defines the function strategy how the model's and it's fields are being named.
	NamerFunc namer.Namer
	// CreateValidator is used as a validator for the Create processes.
	CreateValidator *validator.Validate
	//PatchValidator is used as a validator for the Patch processes.
	PatchValidator *validator.Validate

	// modelMap is a mapping for the model's structures.
	ModelMap *mapping.ModelMap
	// modelRepositories is the mapping of the model's to their repositories.
	modelRepositories map[*mapping.ModelStruct]repository.Repository
}

// MustGetNew creates new controller for given provided 'cfg' config.
// Panics on error.
func MustGetNew(cfg *config.Controller) *Controller {
	c, err := newController(cfg)
	if err != nil {
		panic(err)
	}
	return c

}

// New creates new controller for given config 'cfg'.
func New(cfg *config.Controller) (*Controller, error) {
	c, err := newController(cfg)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Close closes all repository instances.
func (c *Controller) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	return repository.CloseAll(ctx)
}

// GetRepository gets the repository for the provided model.
func (c *Controller) GetRepository(model interface{}) (repository.Repository, error) {
	mStruct, err := c.getModelStruct(model)
	if err != nil {
		return nil, err
	}

	repo, ok := c.modelRepositories[mStruct]
	if !ok {
		if err := c.mapModel(mStruct); err != nil {
			return nil, err
		}
		repo = c.modelRepositories[mStruct]
	}
	return repo, nil
}

// ListModels returns a list of registered models for given controller.
func (c *Controller) ListModels() []*mapping.ModelStruct {
	var models []*mapping.ModelStruct
	for _, model := range c.ModelMap.Models() {
		models = append(models, (*mapping.ModelStruct)(model))
	}
	return models
}

// ModelStruct gets the model struct on the base of the provided model
func (c *Controller) ModelStruct(model interface{}) (*mapping.ModelStruct, error) {
	m, err := c.getModelStruct(model)
	if err != nil {
		return nil, err
	}
	return (*mapping.ModelStruct)(m), nil
}

// MustGetModelStruct gets the model struct from the cached model Map.
// Panics if the model does not exists in the map.
func (c *Controller) MustGetModelStruct(model interface{}) *mapping.ModelStruct {
	mStruct, err := c.getModelStruct(model)
	if err != nil {
		panic(err)
	}
	return mStruct
}

// RegisterModels registers provided models within the context of the provided Controller
func (c *Controller) RegisterModels(models ...interface{}) error {
	if err := c.checkDefaultRepositories(); err != nil {
		log.Errorf("Registering models failed - no default repository set yet. %v", err)
		return err
	}

	log.Debug2f("Registering '%d' models", len(models))
	if err := c.ModelMap.RegisterModels(models...); err != nil {
		return err
	}

	if err := c.Config.MapRepositories(); err != nil {
		log.Debugf("Mapping models to repositories failed. %v", err)
		return err
	}

	for _, mStruct := range c.ModelMap.Models() {
		log.Debug3f("Checking repository for model: %s", mStruct.Collection())
		if _, err := c.GetRepository(mStruct); err != nil {
			log.Errorf("Mapping model: '%v' to repository failed.", mStruct.Type().Name())
			return err
		}
	}
	return nil
}

// RegisterRepository registers provided repository for given 'name' and with with given 'cfg' config.
func (c *Controller) RegisterRepository(name string, cfg *config.Repository) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	// check if the repository is not already registered
	_, ok := c.Config.Repositories[name]
	if ok {
		return errors.NewDetf(class.RepositoryConfigAlreadyRegistered, "repository: '%s' already exists", name)
	}
	c.Config.Repositories[name] = cfg

	// if there is not default repository check if the repository could be default.
	if c.Config.DefaultRepository == nil {
		if c.Config.DefaultRepositoryName == name || c.Config.DefaultRepositoryName == "" {
			log.Infof("Setting default repository to: '%s'", name)
			c.Config.DefaultRepository = cfg
			c.Config.DefaultRepositoryName = name
		}
	}
	return nil
}

func (c *Controller) checkDefaultRepositories() error {
	if c.Config.Repositories == nil {
		return errors.NewDet(class.ConfigValueNil, "no Repositories map found for the controller config")
	}

	if c.Config.DefaultRepository != nil {
		if c.Config.DefaultRepositoryName == "" {
			return errors.NewDetf(class.RepositoryConfigInvalid, "no default repository name provided")
		}
		// check if it is stored in repositories
		_, ok := c.Config.Repositories[c.Config.DefaultRepositoryName]
		if !ok {
			c.Config.Repositories[c.Config.DefaultRepositoryName] = c.Config.DefaultRepository
		}
		return nil
	}

	// find if the repository is in the Repositories
	if c.Config.DefaultRepositoryName == "" && len(c.Config.Repositories) == 0 {
		return errors.NewDet(class.RepositoryNotFound, "no repositories found for the controller")
	}

	if c.Config.DefaultRepositoryName != "" {
		defaultRepo, ok := c.Config.Repositories[c.Config.DefaultRepositoryName]
		if ok {
			c.Config.DefaultRepository = defaultRepo
			return nil
		}
		return errors.NewDetf(class.ConfigValueInvalid, "default repository: '%s' not registered within the controller", c.Config.DefaultRepositoryName)
	}

	return errors.NewDet(class.ConfigValueNil, "no default repository set for the controller")
}

func (c *Controller) getModelStruct(model interface{}) (*mapping.ModelStruct, error) {
	if model == nil {
		return nil, errors.NewDet(class.ModelValueNil, "provided nil model value")
	}

	switch tp := model.(type) {
	case *mapping.ModelStruct:
		return tp, nil
	case string:
		m := c.ModelMap.GetByCollection(tp)
		if m == nil {
			return nil, errors.NewDetf(class.ModelNotMapped, "model: '%s' is not found", tp)
		}
		return m, nil
	}

	mStruct, err := c.ModelMap.GetModelStruct(model)
	if err != nil {
		return nil, err
	}
	return mStruct, nil
}

// setConfig sets and validates provided config
func (c *Controller) setConfig(cfg *config.Controller) error {
	// if there is no controller config provided throw an error.
	if cfg == nil {
		return errors.NewDet(class.ConfigValueNil, "provided nil config value")
	}

	// set the log level from the provided config.
	if cfg.LogLevel != "" {
		level := unilogger.ParseLevel(cfg.LogLevel)
		if level == unilogger.UNKNOWN {
			return errors.NewDetf(class.ConfigValueInvalid, "invalid 'log_level' value: '%s'", cfg.LogLevel)
		}

		if log.Logger() == nil {
			log.Default()
		}

		if log.Level() != level {
			// get and set default logger
			if err := log.SetLevel(level); err != nil {
				return err
			}
		}
	}
	log.Debug2f("Creating new controller with config: '%#v'", cfg)

	// set the naming convention
	cfg.NamingConvention = strings.ToLower(cfg.NamingConvention)

	if err := validate.Struct(cfg); err != nil {
		return errors.NewDet(class.ConfigValueInvalid, "validating config failed")
	}

	if err := cfg.Processor.Validate(); err != nil {
		return err
	}

	// if there is no repositories create a map container
	if cfg.Repositories == nil {
		cfg.Repositories = make(map[string]*config.Repository)
	}

	// iterate over all repositories and find the default one
	if cfg.DefaultRepository == nil {
		for name, repoConfig := range cfg.Repositories {
			if cfg.DefaultRepositoryName == name || cfg.DefaultRepositoryName == "" {
				log.Debugf("Setting default repository to: '%s'", name)
				cfg.DefaultRepository = repoConfig
				cfg.DefaultRepositoryName = name
				break
			}
		}
	}

	// if the default is found check if it is stored within the Repositories
	if cfg.DefaultRepository != nil && cfg.DefaultRepositoryName != "" {
		_, ok := cfg.Repositories[cfg.DefaultRepositoryName]
		if !ok {
			cfg.Repositories[cfg.DefaultRepositoryName] = cfg.DefaultRepository
		}
	} else if cfg.DefaultRepository != nil && cfg.DefaultRepositoryName == "" {
		return errors.NewDet(class.ConfigValueInvalid, "default repository have no name defined in the Controller Config")
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
	default:
		return errors.NewDetf(class.ConfigValueInvalid, "unknown naming convention name: %s", cfg.NamingConvention)
	}

	log.Debugf("Naming Convention used in schemas: %s", cfg.NamingConvention)
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

func (c *Controller) mapModel(model *mapping.ModelStruct) error {
	repoName := model.Config().Repository.DriverName
	if repoName == "" {
		return errors.NewDet(class.ModelSchemaNotFound, "no default repository factory found")
	}

	factory := repository.GetFactory(repoName)
	if factory == nil {
		err := errors.NewDetf(class.RepositoryFactoryNotFound, "repository factory: '%s' not found.", repoName)
		log.Debug(err)
		return err
	}

	repo, err := factory.New(c, (*mapping.ModelStruct)(model))
	if err != nil {
		return err
	}

	c.modelRepositories[model] = repo

	log.Debugf("Model %s mapped, to repository: %s", model.Collection(), repoName)
	return nil
}

func newController(cfg *config.Controller) (*Controller, error) {
	var err error
	c := &Controller{
		CreateValidator:   validator.New(),
		PatchValidator:    validator.New(),
		modelRepositories: make(map[*mapping.ModelStruct]repository.Repository),
	}

	if err = c.setConfig(cfg); err != nil {
		return nil, err
	}

	// create and initialize model mapping
	c.ModelMap = mapping.NewModelMap(c.NamerFunc, c.Config)
	return c, nil
}
