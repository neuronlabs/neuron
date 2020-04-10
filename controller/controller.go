package controller

import (
	"strings"
	"time"

	"gopkg.in/go-playground/validator.v9"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"
	"github.com/neuronlabs/neuron-core/namer"
	"github.com/neuronlabs/neuron-core/repository"
)

var validate = validator.New()

// Controller is the structure that contains, initialize and control the flow of the application.
// It contains repositories, model definitions, validators.
type Controller struct {
	// Config is the configuration struct for the controller.
	Config *config.Controller
	// NamerFunc defines the function strategy how the model's and it's fields are being named.
	NamerFunc namer.Namer
	// CreateValidator is used as a validator for the Create processes.
	CreateValidator *validator.Validate
	// PatchValidator is used as a validator for the Patch processes.
	PatchValidator *validator.Validate

	// ModelMap is a mapping for the model's structures.
	ModelMap *mapping.ModelMap
	// Repositories is the mapping of the repositoryName to the repository.
	Repositories map[string]repository.Repository
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

// GetRepository gets the repository for the provided model.
func (c *Controller) GetRepository(model interface{}) (repository.Repository, error) {
	mStruct, err := c.getModelStruct(model)
	if err != nil {
		return nil, err
	}
	repo, ok := c.Repositories[mStruct.Config().RepositoryName]
	if !ok {
		return nil, errors.NewDetf(class.RepositoryNotFound, "no repository found for the model")
	}
	return repo, nil
}

// ListModels returns a list of registered models for given controller.
func (c *Controller) ListModels() []*mapping.ModelStruct {
	return c.ModelMap.Models()
}

// ModelStruct gets the model struct on the base of the provided model
func (c *Controller) ModelStruct(model interface{}) (*mapping.ModelStruct, error) {
	return c.getModelStruct(model)
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

// RegisterModels registers provided models within the context of the provided Controller.
// All repositories must be registered up to this moment.
func (c *Controller) RegisterModels(models ...interface{}) (err error) {
	log.Debug2f("Registering '%d' models", len(models))
	start := time.Now()
	// register models
	if err = c.ModelMap.RegisterModels(models...); err != nil {
		return err
	}
	// map models to their repositories.
	if err = c.mapAndRegisterInRepositories(); err != nil {
		return err
	}
	log.Debug2f("Models registered in: %s", time.Since(start))
	return nil
}

func (c *Controller) mapAndRegisterInRepositories(models ...interface{}) error {
	// map models to their repositories in the config
	if err := c.Config.MapModelsRepositories(); err != nil {
		return err
	}

	// match models to their repository instances.
	modelsRepositories := make(map[repository.Repository][]*mapping.ModelStruct)
	for _, model := range models {
		mStruct, err := c.getModelStruct(model)
		if err != nil {
			return err
		}
		repo, ok := c.Repositories[mStruct.Config().RepositoryName]
		if !ok {
			return errors.NewDetf(class.RepositoryNotFound, "repository not found for the model: '%s'", mStruct.String())
		}
		modelsRepositories[repo] = append(modelsRepositories[repo], mStruct)
	}
	for repo, modelsStructures := range modelsRepositories {
		if err := repo.RegisterModels(modelsStructures...); err != nil {
			log.Errorf("Registering models in repository: %v failed: %v", repo.FactoryName(), modelsStructures)
			return err
		}
	}
	return nil
}

// RegisterRepository registers provided repository for given 'name' and with with given 'cfg' config.
func (c *Controller) RegisterRepository(name string, cfg *config.Repository) (err error) {
	if err = cfg.Validate(); err != nil {
		return err
	}
	if _, ok := c.Repositories[name]; ok {
		return errors.NewDetf(class.RepositoryConfigAlreadyRegistered, "repository: '%s' already exists", name)
	}
	// check if the repository is already registered
	if err = c.setRepositoryConfig(name, cfg); err != nil {
		return err
	}
	// create new repository from factory.
	repo, err := c.newRepository(cfg)
	if err != nil {
		return err
	}
	// map the repository to its name.
	c.Repositories[name] = repo
	return nil
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
func (c *Controller) setConfig(cfg *config.Controller) (err error) {
	// if there is no controller config provided throw an error.
	if cfg == nil {
		return errors.NewDet(class.ConfigValueNil, "provided nil config value")
	}

	log.Debug3f("Creating new controller with config: '%#v'", cfg)
	// set the naming convention
	cfg.NamingConvention = strings.ToLower(cfg.NamingConvention)

	// map the processor to it's name
	if err = cfg.MapProcessor(); err != nil {
		return err
	}
	// validate config
	if err = validate.Struct(cfg); err != nil {
		return errors.NewDet(class.ConfigValueInvalid, "validating config failed")
	}
	// validate processor functions
	if err = cfg.Processor.Validate(); err != nil {
		return err
	}

	// set naming convention
	switch cfg.NamingConvention {
	case "kebab":
		c.NamerFunc = namer.NamingKebab
	case "camel":
		c.NamerFunc = namer.NamingCamel
	case "lower_camel":
		c.NamerFunc = namer.NamingLowerCamel
	case "snake":
		c.NamerFunc = namer.NamingSnake
	default:
		return errors.NewDetf(class.ConfigValueInvalid, "unknown naming convention name: %s", cfg.NamingConvention)
	}
	log.Debugf("Naming Convention used in schemas: %s", cfg.NamingConvention)

	// set create validation struct tag
	if cfg.CreateValidatorAlias == "" {
		cfg.CreateValidatorAlias = "create"
	}
	c.CreateValidator.SetTagName(cfg.CreateValidatorAlias)
	log.Debug2f("Using: '%s' create validator struct tag", cfg.CreateValidatorAlias)

	// set patch validation struct tag
	if cfg.PatchValidatorAlias == "" {
		cfg.PatchValidatorAlias = "patch"
	}
	c.PatchValidator.SetTagName(cfg.PatchValidatorAlias)
	log.Debug2f("Using: '%s' patch validator struct tag", cfg.PatchValidatorAlias)

	if cfg.Processor == nil {
		cfg.Processor = config.ThreadSafeProcessor()
	}

	c.Config = cfg

	// map repositories from config.
	for name, repositoryConfig := range cfg.Repositories {
		if _, ok := c.Repositories[name]; ok {
			return errors.NewDetf(class.RepositoryConfigInvalid, "repository: '%s' already registered for the controller", name)
		}
		// create new repository from factory.
		repo, err := c.newRepository(repositoryConfig)
		if err != nil {
			return err
		}
		// map the repository to its name.
		c.Repositories[name] = repo
	}
	return nil
}

func (c *Controller) setRepositoryConfig(name string, cfg *config.Repository) error {
	_, ok := c.Config.Repositories[name]
	if ok {
		return errors.NewDetf(class.RepositoryConfigAlreadyRegistered, "repository: '%s' already exists", name)
	}
	c.Config.Repositories[name] = cfg

	// if there is not default repository and the default repository name matches or is not defined
	// set this repository as default
	if c.Config.DefaultRepository == nil && (c.Config.DefaultRepositoryName == name || c.Config.DefaultRepositoryName == "") {
		log.Infof("Setting default repository to: '%s'", name)
		c.Config.DefaultRepository = cfg
		c.Config.DefaultRepositoryName = name
	}
	return nil
}

func (c *Controller) newRepository(cfg *config.Repository) (repository.Repository, error) {
	driverName := cfg.DriverName
	if driverName == "" {
		log.Errorf("No driver name specified for the repository configuration: %v", cfg)
		return nil, errors.NewDetf(class.RepositoryConfigInvalid, "no repository driver name found for the repository: %v", cfg)
	}
	factory := repository.GetFactory(driverName)
	if factory == nil {
		log.Errorf("Factory for driver: '%s' is not found", driverName)
		return nil, errors.NewDetf(class.RepositoryFactoryNotFound, "repository factory: '%s' not found.", driverName)
	}
	repo, err := factory.New(c, cfg)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func newController(cfg *config.Controller) (*Controller, error) {
	var err error
	c := &Controller{
		CreateValidator: validator.New(),
		PatchValidator:  validator.New(),
		Repositories:    make(map[string]repository.Repository),
	}

	if err = c.setConfig(cfg); err != nil {
		return nil, err
	}

	// create and initialize model mapping
	c.ModelMap = mapping.NewModelMap(c.NamerFunc, c.Config)
	return c, nil
}
