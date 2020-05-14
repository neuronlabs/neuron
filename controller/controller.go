package controller

import "C"
import (
	"strings"
	"time"

	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
)

// Controller is the structure that contains, initialize and control the flow of the application.
// It contains repositories, model definitions.
type Controller struct {
	// Config is the configuration struct for the controller.
	Config *config.Controller
	// NamerFunc defines the function strategy how the model's and it's fields are being named.
	NamerFunc mapping.Namer
	// ModelMap is a mapping for the model's structures.
	ModelMap *mapping.ModelMap
	// Repositories is the mapping of the repositoryName to the repository.
	Repositories map[string]repository.Repository
	// DefaultRepository is the default repository or given controller.
	DefaultRepository repository.Repository
}

// NewController creates new controller for given config 'cfg'.
func NewController(cfg *config.Controller) (*Controller, error) {
	c, err := newController(cfg)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// GetRepository gets the repository for the provided model.
func (c *Controller) GetRepository(model mapping.Model) (repository.Repository, error) {
	mStruct, err := c.getModelStruct(model)
	if err != nil {
		return nil, err
	}
	repo, ok := c.Repositories[mStruct.RepositoryName]
	if !ok {
		return nil, errors.Newf(ClassRepositoryNotFound, "repository not found for the model: %s", mStruct)
	}
	return repo, nil
}

// GetRepositoryByStruct gets the repository by model struct.
func (c *Controller) GetRepositoryByStruct(model *mapping.ModelStruct) (repository.Repository, error) {
	repo, ok := c.Repositories[model.RepositoryName]
	if !ok {
		return nil, errors.Newf(ClassRepositoryNotFound, "repository not found for the model: %s", model)
	}
	return repo, nil
}

// RegisterRepository registers provided repository for given 'name' and with with given 'cfg' config.
func (c *Controller) RegisterRepository(name string, cfg *config.Repository) (err error) {
	if err = cfg.Validate(); err != nil {
		return err
	}
	if _, ok := c.Repositories[name]; ok {
		return errors.Newf(ClassRepositoryNotFound, "repository: '%s' already exists", name)
	}
	// check if the repository is already registered
	_, ok := c.Config.Repositories[name]
	if ok {
		return errors.NewDetf(ClassRepositoryAlreadyRegistered, "repository: '%s' already registered", name)
	}
	c.Config.Repositories[name] = cfg
	// create new repository from factory.
	if err = c.newRepository(name, cfg); err != nil {
		return err
	}
	return nil
}

// Now gets and returns current timestamp. If the configs specify the function might return UTC timestamp.
func (c *Controller) Now() time.Time {
	ts := time.Now()
	if c.Config.UTCTimestamps {
		ts = ts.UTC()
	}
	return ts
}

// setConfig sets and validates provided config
func (c *Controller) setConfig(cfg *config.Controller) (err error) {
	// if there is no controller config provided throw an error.
	if cfg == nil {
		return errors.NewDet(ClassInvalidConfig, "provided nil config value")
	}

	log.Debug3f("Creating new controller with config: '%#v'", cfg)
	// set the naming convention
	cfg.NamingConvention = strings.ToLower(cfg.NamingConvention)

	// validate config
	if err = cfg.Validate(); err != nil {
		return errors.NewDet(ClassInvalidConfig, "validating config failed")
	}

	// set naming convention
	switch cfg.NamingConvention {
	case "kebab":
		c.NamerFunc = mapping.NamingKebab
	case "camel":
		c.NamerFunc = mapping.NamingCamel
	case "lower_camel":
		c.NamerFunc = mapping.NamingLowerCamel
	case "snake":
		c.NamerFunc = mapping.NamingSnake
	default:
		return errors.NewDetf(ClassInvalidConfig, "unknown naming convention name: %s", cfg.NamingConvention)
	}
	log.Debugf("Naming Convention used in schemas: %s", cfg.NamingConvention)
	c.Config = cfg

	// Map repositories from config.
	for name, repositoryConfig := range cfg.Repositories {
		if _, ok := c.Repositories[name]; ok {
			return errors.NewDetf(ClassInvalidConfig, "config repository: '%s' already registered for the controller", name)
		}
		// Insert new repository from factory.
		err := c.newRepository("", repositoryConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) newRepository(name string, cfg *config.Repository) error {
	driverName := cfg.DriverName
	if driverName == "" {
		log.Errorf("No driver name specified for the repository configuration: %v", cfg)
		return errors.NewDetf(ClassInvalidConfig, "no repository driver name found for the repository: %v", cfg)
	}
	factory := repository.GetFactory(driverName)
	if factory == nil {
		log.Errorf("Factory for driver: '%s' is not found", driverName)
		return errors.NewDetf(ClassRepositoryNotFound, "repository factory: '%s' not found.", driverName)
	}
	repo, err := factory.New(cfg)
	if err != nil {
		return err
	}
	c.Repositories[name] = repo

	if c.Config.DefaultRepositoryName == name {
		c.DefaultRepository = repo
	}
	return nil
}

func newController(cfg *config.Controller) (*Controller, error) {
	var err error
	c := &Controller{Repositories: make(map[string]repository.Repository)}
	if err = c.setConfig(cfg); err != nil {
		return nil, err
	}

	c.ModelMap = mapping.NewModelMap(c.NamerFunc)
	return c, nil
}
