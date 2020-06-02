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
	"github.com/neuronlabs/neuron/service"
)

// Controller is the structure that contains, initialize and control the flow of the application.
// It contains repositories, model definitions.
type Controller struct {
	// Config is the configuration struct for the controller.
	Config *config.Controller
	// NamingConvention defines the function strategy how the model's and it's fields are being named.
	NamingConvention mapping.NamingConvention
	// ModelMap is a mapping for the model's structures.
	ModelMap *mapping.ModelMap
	// Services is the mapping of the service name to the given service.
	Services     map[string]service.Service
	Repositories map[string]repository.Repository
	// DefaultService is the default service or given controller.
	DefaultRepository repository.Repository
	defaultRepository string
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
	svc, ok := c.Services[mStruct.RepositoryName]
	if !ok {
		return nil, errors.Newf(ClassRepositoryNotFound, "service not found for the model: %s", mStruct)
	}
	repo, ok := svc.(repository.Repository)
	if !ok {
		return nil, errors.New(ClassRepositoryNotMatched, "provided service is not a repository")
	}
	return repo, nil
}

// GetRepositoryByModelStruct gets the service by model struct.
func (c *Controller) GetRepositoryByModelStruct(model *mapping.ModelStruct) (repository.Repository, error) {
	svc, ok := c.Services[model.RepositoryName]
	if !ok {
		return nil, errors.Newf(ClassRepositoryNotFound, "service not found for the model: %s", model)
	}
	repo, ok := svc.(repository.Repository)
	if !ok {
		return nil, errors.New(ClassRepositoryNotMatched, "provided service is not a repository")
	}
	return repo, nil
}

// RegisterService registers provided service for given 'name' and with with given 'cfg' config.
func (c *Controller) RegisterService(name string, cfg *config.Service) (err error) {
	if err = cfg.Validate(); err != nil {
		return err
	}
	if _, ok := c.Services[name]; ok {
		return errors.Newf(ClassRepositoryNotFound, "service: '%s' already exists", name)
	}
	// Check if the service is already registered.
	_, ok := c.Config.Services[name]
	if ok {
		return errors.NewDetf(ClassRepositoryAlreadyRegistered, "service: '%s' already registered", name)
	}
	c.Config.Services[name] = cfg
	// Create new service from factory.
	if err = c.newService(name, cfg); err != nil {
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

	if err = c.NamingConvention.Parse(cfg.NamingConvention); err != nil {
		return err
	}
	log.Debugf("Naming Convention used in schemas: %s", cfg.NamingConvention)
	c.Config = cfg

	// Map repositories from config.
	for name, serviceConfig := range cfg.Services {
		if _, ok := c.Services[name]; ok {
			return errors.NewDetf(ClassInvalidConfig, "config service: '%s' already registered for the controller", name)
		}
		// Insert new service from factory.
		err := c.newService("", serviceConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) newService(name string, cfg *config.Service) error {
	driverName := cfg.DriverName
	if driverName == "" {
		log.Errorf("No driver name specified for the service configuration: %v", cfg)
		return errors.NewDetf(ClassInvalidConfig, "no service driver name found for the service: %v", cfg)
	}
	factory := service.GetFactory(driverName)
	if factory == nil {
		log.Errorf("Factory for driver: '%s' is not found", driverName)
		return errors.NewDetf(ClassRepositoryNotFound, "service factory: '%s' not found.", driverName)
	}
	svc, err := factory.New(cfg)
	if err != nil {
		return err
	}
	c.Services[name] = svc
	repo, ok := svc.(repository.Repository)
	if !ok {
		return nil
	}
	if c.DefaultRepository == nil {
		c.DefaultRepository = repo
		c.defaultRepository = name
	}

	if c.Config.DefaultRepositoryName == name {
		c.DefaultRepository = repo
		c.defaultRepository = name
	}
	return nil
}

func newController(cfg *config.Controller) (*Controller, error) {
	var err error
	c := &Controller{Services: make(map[string]service.Service)}
	if err = c.setConfig(cfg); err != nil {
		return nil, err
	}

	c.ModelMap = mapping.NewModelMap(c.NamingConvention)
	return c, nil
}
