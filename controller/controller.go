package controller

import "C"
import (
	"strings"
	"time"

	"github.com/neuronlabs/neuron/class"
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/namer"
	"github.com/neuronlabs/neuron/repository"
)

// Controller is the structure that contains, initialize and control the flow of the application.
// It contains repositories, model definitions.
type Controller struct {
	// Config is the configuration struct for the controller.
	Config *config.Controller
	// NamerFunc defines the function strategy how the model's and it's fields are being named.
	NamerFunc namer.Namer
	// ModelMap is a mapping for the model's structures.
	ModelMap *mapping.ModelMap
	// Repositories is the mapping of the repositoryName to the repository.
	Repositories map[string]repository.Repository
}

// MustNewController creates new controller for given provided 'cfg' config.
// Panics on error.
func MustNewController(cfg *config.Controller) *Controller {
	c, err := newController(cfg)
	if err != nil {
		panic(err)
	}
	return c
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
func (c *Controller) GetRepository(model interface{}) (repository.Repository, error) {
	mStruct, err := c.getModelStruct(model)
	if err != nil {
		return nil, err
	}
	repo, ok := c.Repositories[mStruct.Config().RepositoryName]
	if !ok {
		return nil, errors.Newf(ClassRepositoryNotFound, "repository not found for the model: %s", mStruct)
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
		c.NamerFunc = namer.NamingKebab
	case "camel":
		c.NamerFunc = namer.NamingCamel
	case "lower_camel":
		c.NamerFunc = namer.NamingLowerCamel
	case "snake":
		c.NamerFunc = namer.NamingSnake
	default:
		return errors.NewDetf(ClassInvalidConfig, "unknown naming convention name: %s", cfg.NamingConvention)
	}
	log.Debugf("Naming Convention used in schemas: %s", cfg.NamingConvention)
	c.Config = cfg

	if cfg.DefaultRepositoryName != "" && cfg.DefaultRepository != nil {
		if _, ok := cfg.Repositories[cfg.DefaultRepositoryName]; !ok {
			cfg.Repositories[cfg.DefaultRepositoryName] = cfg.DefaultRepository
		}
	}

	// Map repositories from config.
	for name, repositoryConfig := range cfg.Repositories {
		if _, ok := c.Repositories[name]; ok {
			return errors.NewDetf(ClassInvalidConfig, "config repository: '%s' already registered for the controller", name)
		}
		// Insert new repository from factory.
		repo, err := c.newRepository(repositoryConfig)
		if err != nil {
			return err
		}
		// Map the repository to its name.
		c.Repositories[name] = repo
	}
	return nil
}

func (c *Controller) setRepositoryConfig(name string, cfg *config.Repository) error {
	_, ok := c.Config.Repositories[name]
	if ok {
		return errors.NewDetf(ClassRepositoryAlreadyRegistered, "repository: '%s' already registered", name)
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
	repo, err := factory.New(cfg)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func newController(cfg *config.Controller) (*Controller, error) {
	var err error
	c := &Controller{Repositories: make(map[string]repository.Repository)}
	if err = c.setConfig(cfg); err != nil {
		return nil, err
	}

	c.ModelMap = mapping.NewModelMap(c.NamerFunc, c.Config)
	return c, nil
}
