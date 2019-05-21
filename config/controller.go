package config

import (
	"errors"
	"fmt"
	"github.com/neuronlabs/neuron/log"
)

// ControllerConfig defines the configuration for the Controller
type ControllerConfig struct {

	// NamingConvention is the naming convention used while preparing the models.
	// Allowed values:
	// - camel
	// - lowercamel
	// - snake
	// - kebab
	NamingConvention string `mapstructure:"naming_convention" validate:"isdefault|oneof=camel lowercamel snake kebab"`

	// DefaultSchema is the default schema name for the models within given controller
	DefaultSchema string `validate:"alphanum" mapstructure:"default_schema"`

	// ModelSchemas defines the model schemas used by api
	ModelSchemas map[string]*Schema `mapstructure:"schemas"`

	// StrictUnmarshalMode is the flag that defines if the unmarshaling should be in a
	// strict mode that checks if incoming values are all known to the controller
	// As well as the query builder doesn't allow unknown queries
	StrictUnmarshalMode bool `mapstructure:"strict_unmarshal"`

	// Debug sets the debug mode for the controller.
	Debug bool `mapstructure:"debug"`

	// Builder defines the builder config
	Builder *BuilderConfig `mapstructure:"builder"`

	// I18n defines i18n config
	I18n *I18nConfig `mapstructure:"i18n"`

	// Flags defines the controller default flags
	Flags *Flags `mapstructure:"flags"`

	// Repositories contains the connection configs for the given repository instance name
	Repositories map[string]*Repository `mapstructure:"repositories" validate:"-"`

	// DefaultRepositoryName defines default repositoy name
	DefaultRepositoryName string `mapstructure:"default_repository_name"`

	// DefaultRepository defines controller default repository
	DefaultRepository *Repository `mapstructure:"default_repository" validate:"-"`

	// CreateValidatorAlias is the alias for the create validators
	CreateValidatorAlias string `mapstructure:"create_validator_alias"`

	// PatchValidatorAlias is the alis used for the Patch validator
	PatchValidatorAlias string `mapstructure:"patch_validator_alias"`

	// DefaultValidatorAlias is the alias used as a default validator alias
	DefaultValidatorAlias string `mapstructure:"default_validator_alias"`
}

// MapRepositories maps the repositories definitions from the controller with the model's repositories
func (c *ControllerConfig) MapRepositories(s *Schema) error {
	if c.Repositories == nil {
		return errors.New("No repositories found within the config")
	}
	for _, model := range s.Models {
		if model.Repository == nil {
			var reponame string
			if reponame = model.RepositoryName; reponame == "" {
				log.Debugf("Model: %s config have no Repository nor RepositoryName defined. Setting to default repository", model.Collection)
				reponame = c.DefaultRepositoryName
			}
			repoConfig, ok := c.Repositories[reponame]
			if !ok {
				return fmt.Errorf("No repository config definition found for the repository: %s", reponame)
			}
			model.Repository = repoConfig
		}
	}
	return nil
}

// SetDefaultRepository sets the default repository if defined
func (c *ControllerConfig) SetDefaultRepository() error {
	if c.DefaultRepository != nil && c.DefaultRepositoryName != "" {
		if c.Repositories == nil {
			c.Repositories = map[string]*Repository{}
		}
		c.Repositories[c.DefaultRepositoryName] = c.DefaultRepository
	} else if repoName := c.DefaultRepositoryName; repoName != "" && len(c.Repositories) > 0 {

		repo, ok := c.Repositories[repoName]
		if !ok {
			return fmt.Errorf("Default repository: %s not defined in the ControllerConfig.Repository map", repoName)
		}
		c.DefaultRepository = repo

	} else if len(c.Repositories) == 1 {

		for repoName, repo := range c.Repositories {
			c.DefaultRepositoryName = repoName
			c.DefaultRepository = repo
		}
	} else {
		return errors.New("No repositories found within the config")

	}

	return nil
}
