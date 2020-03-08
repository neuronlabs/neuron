package config

import (
	"fmt"

	"github.com/neuronlabs/errors"
	unilogger "github.com/neuronlabs/uni-logger"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
)

// Controller defines the configuration for the Controller.
type Controller struct {
	// NamingConvention is the naming convention used while preparing the models.
	// Allowed values:
	// - camel
	// - lowercamel
	// - snake
	// - kebab
	NamingConvention string `mapstructure:"naming_convention" validate:"isdefault|oneof=camel lowercamel snake kebab"`
	// Models defines the model's configurations.
	Models map[string]*ModelConfig `mapstructure:"models"`
	// Repositories contains the connection configs for the given repository instance name
	Repositories map[string]*Repository `mapstructure:"repositories" validate:"-"`

	// LogLevel is the current logging level
	LogLevel string `mapstructure:"log_level" validate:"isdefault|oneof=debug3 debug2 debug info warning error critical"`

	// DefaultRepositoryName defines default repositoy name
	DefaultRepositoryName string `mapstructure:"default_repository_name"`
	// DefaultRepository defines controller default repository
	DefaultRepository *Repository `mapstructure:"default_repository" validate:"-"`
	// DisallowDefaultRepository determines if the default repository are allowed.
	DisallowDefaultRepository bool `mapstructure:"disallow_default_repository"`

	// Processor is the config used for the scope processor
	ProcessorName string     `mapstructure:"processor_name"`
	Processor     *Processor `mapstructure:"processor" validate:"required"`

	// CreateValidatorAlias is the alias for the create validators
	CreateValidatorAlias string `mapstructure:"create_validator_alias"`
	// PatchValidatorAlias is the alis used for the Patch validator
	PatchValidatorAlias string `mapstructure:"patch_validator_alias"`
	// DefaultValidatorAlias is the alias used as a default validator alias
	DefaultValidatorAlias string `mapstructure:"default_validator_alias"`

	// IncludedDepthLimit is the default limit for the nested included query paremeter.
	// Each model can specify its custom limit in the model's config.
	IncludedDepthLimit int `mapstructure:"included_depth_limit"`
}

// MapModelsRepositories maps the repositories configurations for all model's.
func (c *Controller) MapModelsRepositories() error {
	log.Debug3("Mapping repositories to model configs")
	// iterate over all models in the config and map their repository names to the
	// config's repositories
	for _, model := range c.Models {
		if model.Repository != nil {
			// skip if the repository is already set
			continue
		}
		log.Debug2f("Mapping repository for model: '%s'", model.Collection)

		if model.RepositoryName == "" {
			log.Debug2f("Model: %s config have no Repository nor RepositoryName defined", model.Collection)
			// check if default repositories are allowed
			if c.DisallowDefaultRepository {
				log.Errorf("No repository defined for the model: '%s'. Default repositories are disallowed", model.Collection)
				return fmt.Errorf("no repository config definition found for the: '%s' repository name", model.RepositoryName)
			}
			log.Debug2f("Setting default repository for model: '%s'", model.Collection)
			// set default repository
			model.RepositoryName = c.DefaultRepositoryName
		}

		// check if the repository is defined within the config.
		repoConfig, ok := c.Repositories[model.RepositoryName]
		if !ok {
			log.Errorf("Repository: '%s' not found within configuration repositories: %v", model.RepositoryName, c.Repositories)
			return fmt.Errorf("no repository config definition found for the: '%s' repository name", model.RepositoryName)
		}
		model.Repository = repoConfig
		log.Debugf("Mapped repository: %s for the model: %s", model.RepositoryName, model.Collection)
	}
	return nil
}

// MapProcessor if the processor name is provided maps it to registered processors.
// Returns error if the processor is not defined.
func (c *Controller) MapProcessor() error {
	return c.mapProcessor()
}

// SetDefaultRepository sets the default repository for given controller.
func (c *Controller) SetDefaultRepository() error {
	if c.DisallowDefaultRepository {
		return nil
	}

	if c.Repositories == nil && c.DefaultRepository == nil {
		return errors.NewDet(class.ConfigValueNil, "no Repositories map found for the controller config")
	}

	if c.DefaultRepository != nil {
		// the default repository is not nil - store it's value in repositories
		if c.DefaultRepositoryName == "" {
			c.DefaultRepositoryName = c.DefaultRepository.DriverName
		}
		if c.DefaultRepositoryName == "" {
			return errors.NewDetf(class.RepositoryConfigInvalid, "the default repository name and the driver name are not provided: '%T'", c.DefaultRepository)
		}
		// check if the repository is already initied
		if c.Repositories == nil {
			c.Repositories = make(map[string]*Repository)
		}

		// check if it is stored in repositories
		_, ok := c.Repositories[c.DefaultRepositoryName]
		if !ok {
			c.Repositories[c.DefaultRepositoryName] = c.DefaultRepository
		}
		return nil
	}

	// if there is no default repository name and no repositories are
	if c.DefaultRepositoryName == "" && len(c.Repositories) == 0 {
		return errors.NewDet(class.RepositoryNotFound, "no registered repositories found for the controller")
	}
	// if no default repository name is set
	// get first orccurance
	if c.DefaultRepositoryName == "" {
		for name := range c.Repositories {
			c.DefaultRepositoryName = name
			break
		}
	}

	defaultRepo, ok := c.Repositories[c.DefaultRepositoryName]
	if ok {
		c.DefaultRepository = defaultRepo
		return nil
	}
	return errors.NewDetf(class.ConfigValueInvalid, "default repository: '%s' not registered within the controller", c.DefaultRepositoryName)
}

// SetLogger sets the logger based on the config log level.
func (c *Controller) SetLogger() error {
	level := unilogger.ParseLevel(c.LogLevel)
	if level == unilogger.UNKNOWN {
		return errors.NewDetf(class.ConfigValueInvalid, "invalid 'log_level' value: '%s'", c.LogLevel)
	}

	// if the logger is null set it to default
	if log.Logger() == nil {
		log.Default()
	}

	if log.Level() != level {
		// get and set default logger
		if err := log.SetLevel(level); err != nil {
			return err
		}
	}
	return nil
}

// mapProcessor gets the processor if the ProcessorName is already set.
func (c *Controller) mapProcessor() error {
	if c.ProcessorName != "" {
		var ok bool
		c.Processor, ok = processors[c.ProcessorName]
		if !ok {
			return errors.NewDetf(class.ConfigValueProcessor, "processor: '%s' not registered", c.ProcessorName)
		}
	}
	if c.Processor == nil {
		return errors.NewDet(class.ConfigValueProcessor, "processor not defined")
	}
	return nil
}
