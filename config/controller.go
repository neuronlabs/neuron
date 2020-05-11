package config

import (
	"fmt"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
)

// Controller defines the configuration for the Controller.
type Controller struct {
	// NamingConvention is the naming convention used while preparing the models.
	// Allowed values:
	// - camel
	// - lower_camel
	// - snake
	// - kebab
	NamingConvention string `mapstructure:"naming_convention"`
	// Models defines the model's configurations.
	Models map[string]*ModelConfig `mapstructure:"models"`
	// Repositories contains the connection configs for the given repository instance name
	Repositories map[string]*Repository `mapstructure:"repositories"`
	// DefaultRepositoryName defines default repository name
	DefaultRepositoryName string `mapstructure:"default_repository_name"`
	// DefaultRepository defines controller default repository
	DefaultRepository *Repository `mapstructure:"default_repository"`
	// DisallowDefaultRepository determines if the default repository are allowed.
	DisallowDefaultRepository bool `mapstructure:"disallow_default_repository"`
	// AsynchronousIncludes defines if the query relation includes would be taken concurrently.
	AsynchronousIncludes bool `mapstructure:"asynchronous_includes"`
	// UTCTimestamps is the flag that defines the format of the timestamps.
	UTCTimestamps bool `mapstructure:"utc_timestamps"`
}

// Validate checks the validity of the config.
func (c *Controller) Validate() error {
	if c.NamingConvention != "" {
		var found bool
		for _, naming := range []string{"camel", "lower_camel", "snake", "kebab"} {
			if c.NamingConvention == naming {
				found = true
				break
			}
		}
		if !found {
			return errors.Newf(ClassConfigInvalidValue, "provided invalid naming convention")
		}
	}
	return nil
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
				log.Errorf("No repository defined for the model: '%s'. DefaultController repositories are disallowed", model.Collection)
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

// SetDefaultRepository sets the default repository for given controller.
func (c *Controller) SetDefaultRepository() error {
	if c.DisallowDefaultRepository {
		return nil
	}

	if c.Repositories == nil && c.DefaultRepository == nil {
		return errors.NewDet(ClassConfigInvalidValue, "no Repositories map found for the controller config")
	}

	if c.DefaultRepository != nil {
		// the default repository is not nil - store it's value in repositories
		if c.DefaultRepositoryName == "" {
			c.DefaultRepositoryName = c.DefaultRepository.DriverName
		}
		if c.DefaultRepositoryName == "" {
			return errors.NewDetf(ClassConfigInvalidValue, "the default repository name and the driver name are not provided: '%T'", c.DefaultRepository)
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
		return errors.NewDet(ClassConfigInvalidValue, "no registered repositories found for the controller")
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
	return errors.NewDetf(ClassConfigInvalidValue, "default repository: '%s' not registered within the controller", c.DefaultRepositoryName)
}
