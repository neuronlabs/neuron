package config

import (
	"fmt"

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

	// StrictUnmarshalMode is the flag that defines if the unmarshaling should be in a
	// strict mode that checks if incoming values are all known to the controller
	// As well as the query builder doesn't allow unknown queries
	StrictUnmarshalMode bool `mapstructure:"strict_unmarshal"`

	// EncodeLinks is the boolean used for encoding the links in the jsonapi encoder
	EncodeLinks bool `mapstructure:"encode_links"`

	// LogLevel is the current logging level
	LogLevel string `mapstructure:"log_level" validate:"isdefault|oneof=debug3 debug2 debug info warning error critical"`

	// Repositories contains the connection configs for the given repository instance name
	Repositories map[string]*Repository `mapstructure:"repositories" validate:"-"`

	// DefaultRepositoryName defines default repositoy name
	DefaultRepositoryName string `mapstructure:"default_repository_name"`

	// DefaultRepository defines controller default repository
	DefaultRepository *Repository `mapstructure:"default_repository" validate:"-"`

	// Processor is the config used for the scope processor
	Processor *Processor `mapstructure:"processor" validate:"required"`

	// CreateValidatorAlias is the alias for the create validators
	CreateValidatorAlias string `mapstructure:"create_validator_alias"`

	// PatchValidatorAlias is the alis used for the Patch validator
	PatchValidatorAlias string `mapstructure:"patch_validator_alias"`

	// DefaultValidatorAlias is the alias used as a default validator alias
	DefaultValidatorAlias string `mapstructure:"default_validator_alias"`
}

// MapRepositories maps the repositories definitions from the controller with the model's repositories.
func (c *Controller) MapRepositories() error {
	log.Debug3("Mapping repositories to model configs")
	for _, model := range c.Models {
		if model.Repository != nil {
			continue
		}
		log.Debug2f("Mapping repository for model: '%s'", model.Collection)

		var reponame string
		if reponame = model.RepositoryName; reponame == "" {
			log.Debugf("Model: %s config have no Repository nor RepositoryName defined. Setting to default repository", model.Collection)
			reponame = c.DefaultRepositoryName
		}

		repoConfig, ok := c.Repositories[reponame]
		if !ok {
			log.Errorf("Repository: '%s' not found within configuration repositories: %v", reponame, c.Repositories)
			return fmt.Errorf("no repository config definition found for the: '%s' repository name", reponame)
		}
		model.Repository = repoConfig
		log.Debugf("Mapped repositpory: %s for the model: %s", reponame, model.Collection)
	}

	return nil
}
